package notify

import (
	"encoding/json"
	"fmt"
	"github.com/analogj/lodestone-publisher/pkg/model"
	"github.com/streadway/amqp"
)

type AmqpNotify struct {
	client   *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

func (n *AmqpNotify) Init(config map[string]string) error {
	n.exchange = config["exchange"]
	n.queue = config["queue"]

	client, err := amqp.Dial(config["amqp-url"])
	if err != nil {
		return err
	}
	n.client = client

	//test connection
	ch, err := client.Channel()
	if err != nil {
		return err
	}
	n.channel = ch

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

	return err
}

func (n *AmqpNotify) Publish(event model.S3Event) error {
	fmt.Println("Publishing event..")

	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return n.channel.Publish(
		n.exchange, // exchange
		n.queue,    // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		},
	)
}

func (n *AmqpNotify) Close() error {
	if err := n.client.Close(); err != nil {
		return err
	}
	if err := n.channel.Close(); err != nil {
		return err
	}
	return nil
}
