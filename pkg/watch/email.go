package watch

import (
	"crypto/tls"
	"fmt"
	"github.com/analogj/lodestone-publisher/pkg/notify"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"log"
	"math"
)

type EmailWatcher struct{}

func (fs *EmailWatcher) Start(notifyClient notify.Interface, config map[string]string) {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(fmt.Sprintf("%s:%s", config["imap-hostname"], config["imap-port"]), &tls.Config{ServerName: config["imap-hostname"]})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(config["imap-username"], config["imap-password"]); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get all messages
	if mbox.Messages == 0 {
		return //nothing to process
	}

	paginatedProcessMessages(c, mbox)

}

func paginatedProcessMessages(c *client.Client, mbox *imap.MailboxStatus) {
	//retrieve messages from mailbox
	//note: the number of messages may be absurdly large, so lets paginate for safety (sets of 100 messages)

	pages := math.Ceil(float64(mbox.Messages) / float64(100))

	for page := 0; page < int(pages); page++ {
		//for each "page", lets generate the range of messages to retrieve

		// message ranges are 1 base indexed.
		// ie, page 1 is messages 1-100
		// page 2 is messages 101-200

		from := uint32(page*100 + 1)
		to := uint32((page + 1) * 100)
		if mbox.Messages < to {
			to = mbox.Messages
		}

		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		log.Printf("Retrieving messages (%v-%v)", from, to)
		retrieveMessages(c, seqset)

	}

}

func retrieveMessages(c *client.Client, seqset *imap.SeqSet) {

	messages := make(chan *imap.Message, 100)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)

		//process the message (download the attachment)

	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Retrieved message set!")
}

func downloadAttachments(c *client.Client) {

}

func deleteMessages(c *client.Client, seqset *imap.SeqSet) {
	// Mark the messages as deleted
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.Store(seqset, item, flags, nil); err != nil {
		log.Fatal(err)
	}

	// Then delete it
	if err := c.Expunge(nil); err != nil {
		log.Fatal(err)
	}
}
