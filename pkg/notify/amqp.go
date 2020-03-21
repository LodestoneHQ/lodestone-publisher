package notify

import (
	"encoding/json"
	"errors"
	"github.com/analogj/lodestone-publisher/pkg/model"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

//Based on https://github.com/streadway/amqp/blob/master/example_client_test.go

type AmqpNotify struct {
	logger   *logrus.Entry
	client   *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string

	done            chan bool
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	isReady         bool
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When setting up the channel after a channel exception
	reInitDelay = 2 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to a server")
	errAlreadyClosed = errors.New("already closed: not connected to the server")
	errShutdown      = errors.New("session is shutting down")
)

func (n *AmqpNotify) Init(logger *logrus.Entry, config map[string]string) error {
	n.exchange = config["exchange"]
	n.queue = config["queue"]
	n.logger = logger

	go n.handleReconnect(config["amqp-url"])
	return nil
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (n *AmqpNotify) handleReconnect(addr string) {
	for {
		n.isReady = false
		n.logger.Infoln("Attempting to connect")

		conn, err := n.connect(addr)

		if err != nil {
			n.logger.Errorln("Failed to connect. Retrying...")

			select {
			case <-n.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := n.handleReInit(conn); done {
			break
		}
	}
}

// connect will create a new AMQP connection
func (n *AmqpNotify) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)

	if err != nil {
		return nil, err
	}

	n.changeClient(conn)
	n.logger.Infoln("Connected!")
	return conn, nil
}

// handleReconnect will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (n *AmqpNotify) handleReInit(conn *amqp.Connection) bool {
	for {
		n.isReady = false

		err := n.init(conn)

		if err != nil {
			n.logger.Warnln("Failed to initialize channel. Retrying...")

			select {
			case <-n.done:
				return true
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-n.done:
			return true
		case <-n.notifyConnClose:
			n.logger.Warnln("Connection closed. Reconnecting...")
			return false
		case <-n.notifyChanClose:
			n.logger.Warnln("Channel closed. Re-running init...")
		}
	}
}

// init will initialize channel & declare queue
func (n *AmqpNotify) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	err = ch.Confirm(false)

	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		n.exchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		n.queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	n.changeChannel(ch)
	n.isReady = true
	n.logger.Debugln("Setup!")

	return nil
}

// changeClient takes a new connection to the queue,
// and updates the close listener to reflect this.
func (n *AmqpNotify) changeClient(client *amqp.Connection) {
	n.client = client
	n.notifyConnClose = make(chan *amqp.Error)
	n.client.NotifyClose(n.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (n *AmqpNotify) changeChannel(channel *amqp.Channel) {
	n.channel = channel
	n.notifyChanClose = make(chan *amqp.Error)
	n.notifyConfirm = make(chan amqp.Confirmation, 1)
	n.channel.NotifyClose(n.notifyChanClose)
	n.channel.NotifyPublish(n.notifyConfirm)
}

// Publish will push data onto the queue, and wait for a confirm.
// If no confirms are received until within the resendTimeout,
// it continuously re-sends messages until a confirm is received.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (n *AmqpNotify) Publish(event model.S3Event) error {
	if !n.isReady {
		return errors.New("failed to publish event: not connected")
	}

	n.logger.Println("Publishing event..")

	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	for {
		err := n.unsafePublish(b)
		if err != nil {
			n.logger.Println("Publish failed. Retrying...")
			select {
			case <-n.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}
		select {
		case confirm := <-n.notifyConfirm:
			if confirm.Ack {
				n.logger.Println("Publish confirmed!")
				return nil
			}
		case <-time.After(resendDelay):
		}
		n.logger.Println("Publish didn't confirm. Retrying...")
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// recieve the message.
func (n *AmqpNotify) unsafePublish(data []byte) error {
	if !n.isReady {
		return errNotConnected
	}
	return n.channel.Publish(
		n.exchange, // exchange
		n.queue,    // routing key
		false,      // Mandatory
		false,      // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
}

func (n *AmqpNotify) Close() error {
	if !n.isReady {
		return errAlreadyClosed
	}
	err := n.channel.Close()
	if err != nil {
		return err
	}
	err = n.client.Close()
	if err != nil {
		return err
	}
	close(n.done)
	n.isReady = false
	return nil
}
