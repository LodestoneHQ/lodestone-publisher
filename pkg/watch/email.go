package watch

import (
	"crypto/tls"
	"fmt"
	"github.com/analogj/lodestone-publisher/pkg/notify"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type EmailWatcher struct {
	apiEndpoint string
	bucket      string
}

func (ew *EmailWatcher) Start(notifyClient notify.Interface, config map[string]string) {

	ew.apiEndpoint = config["api-endpoint"]
	ew.bucket = config["bucket"]

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

	ew.batchProcessMessages(c)

}

func (ew *EmailWatcher) batchProcessMessages(c *client.Client) {
	//retrieve messages from mailbox
	//note: the number of messages may be absurdly large, so lets do this in batches for safety (sets of 100 messages)

	//for each "page", lets generate the range of messages to retrieve

	// message ranges are 1 base indexed.
	// ie, batches include messages from 1-100

	for {
		// get lastest mailbox information
		mbox, err := c.Select("INBOX", false)
		if err != nil {
			log.Fatal(err)
		}
		// Get all messages
		if mbox.Messages == 0 {
			//if theres no messages to process, break out of the loop and wait for next imap interval
			log.Printf("No messages to process")
			break
		}

		from := uint32(1)
		to := uint32(100)
		if mbox.Messages < to {
			to = mbox.Messages
		}

		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		log.Printf("Retrieving messages")
		ew.retrieveMessages(c, seqset)

		//todo publish/generate events for stored documents
		ew.generateEvent()

		//todo, delete messages
		//deleteMessages(c, seqset)
	}

}

func (ew *EmailWatcher) retrieveMessages(c *client.Client, seqset *imap.SeqSet) {
	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchUid}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	for msg := range messages {
		log.Println("UID: ", msg.Uid)
		/* read and process the email */

		ew.storeAttachments(c, section, msg)

	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}

func (ew *EmailWatcher) storeAttachments(c *client.Client, section *imap.BodySectionName, msg *imap.Message) ([]string, error) {
	r := msg.GetBody(section)
	if r == nil {
		log.Print("Error: Message body empty.")
		return nil, nil
	}

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Printf("Error creating mail readerr: %v", err)
		return nil, err
	}

	// Print some info about the message
	header := mr.Header
	if date, err := header.Date(); err == nil {
		log.Println("Date:", date)
	}
	if from, err := header.AddressList("From"); err == nil {
		log.Println("From:", from)
	}
	if to, err := header.AddressList("To"); err == nil {
		log.Println("To:", to)
	}
	if subject, err := header.Subject(); err == nil {
		log.Println("Subject:", subject)
	}

	//TODO: filter message based on sender, attachment type

	//make a temporary directory for subsequent processing (attachment file download)
	localTempDir, err := ioutil.TempDir("", "attach")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(localTempDir) // clean up

	storagePaths := []string{}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		switch h := p.Header.(type) {
		case *mail.AttachmentHeader:
			// This is an attachment
			attachmentFilename, _ := h.Filename()

			localPath, err := ew.saveAttachment(attachmentFilename, p.Body, localTempDir)
			if err != nil {
				continue
			}
			storagePath := filepath.Join("email", attachmentFilename)
			err = ew.uploadAttachmentToStorage(storagePath, localPath)
			if err != nil {
				continue
			}
			storagePaths = append(storagePaths, storagePath)
		}
	}

	return storagePaths, nil
}

func (ew *EmailWatcher) saveAttachment(attachmentFilename string, attachmentData io.Reader, localTempDir string) (string, error) {

	fileName := filepath.Base(attachmentFilename)
	localFilepath := filepath.Join(localTempDir, fileName)
	log.Println("Store attachment locally: %v, %v", attachmentFilename, localFilepath)

	localFile, err := os.Create(localFilepath)
	if err != nil {
		return "", err
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, attachmentData)
	if err != nil {
		return "", err
	}

	return localFilepath, err
}

func (ew *EmailWatcher) uploadAttachmentToStorage(storagePath string, localFilepath string) error {

	localFile, err := os.Open(localFilepath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	//manipulate the path
	apiEndpoint, err := url.Parse(ew.apiEndpoint)
	if err != nil {
		return err
	}
	apiEndpoint.Path = fmt.Sprintf("/api/v1/storage/%s/%s", ew.bucket, storagePath)

	resp, err := http.Post(apiEndpoint.String(), "binary/octet-stream", localFile)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (ew *EmailWatcher) generateEvent() {

}

func (ew *EmailWatcher) deleteMessages(c *client.Client, seqset *imap.SeqSet) {
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
