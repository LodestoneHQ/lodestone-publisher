package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/analogj/lodestone-publisher/pkg/notify"
	"github.com/analogj/lodestone-publisher/pkg/version"
	"github.com/analogj/lodestone-publisher/pkg/watch"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "lodestone-email-publisher",
		Usage:    "Email watcher & notifications for lodestone",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			capsuleUrl := "AnalogJ/lodestone-publisher"

			versionInfo := fmt.Sprintf("%s.%s-%s", goos, goarch, version.VERSION)

			subtitle := capsuleUrl + utils.LeftPad2Len(versionInfo, " ", 53-len(capsuleUrl))

			fmt.Fprintf(c.App.Writer, fmt.Sprintf(utils.StripIndent(
				`
			 __    _____  ____  ____  ___  ____  _____  _  _  ____ 
			(  )  (  _  )(  _ \( ___)/ __)(_  _)(  _  )( \( )( ___)
			 )(__  )(_)(  )(_) ))__) \__ \  )(   )(_)(  )  (  )__) 
			(____)(_____)(____/(____)(___/ (__) (_____)(_)\_)(____)
			%s
			`), subtitle))
			return nil
		},

		Commands: []cli.Command{
			{
				Name:  "start",
				Usage: "Start the Lodestone email watcher",
				Action: func(c *cli.Context) error {

					var notifyClient notify.Interface

					notifyClient = new(notify.AmqpNotify)
					err := notifyClient.Init(map[string]string{
						"amqp-url": c.String("amqp-url"),
						"exchange": c.String("amqp-exchange"),
						"queue":    c.String("amqp-queue"),
					})
					if err != nil {
						return err
					}
					defer notifyClient.Close()

					watcher := watch.EmailWatcher{}
					watcher.Start(notifyClient, map[string]string{
						"imap-hostname": c.String("imap-hostname"),
						"imap-port":     c.String("imap-port"),
						"imap-username": c.String("imap-username"),
						"imap-password": c.String("imap-password"),
						"imap-interval": c.String("imap-interval"),
						"bucket":        c.String("bucket"),
						"api-endpoint":  c.String("api-endpoint"),
					})
					return nil
				},

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "imap-hostname",
						Usage: "The imap server hostname",
					},
					&cli.StringFlag{
						Name:  "imap-port",
						Usage: "The imap server port",
						Value: "993",
					},
					&cli.StringFlag{
						Name:  "imap-username",
						Usage: "The imap server username",
					},
					&cli.StringFlag{
						Name:  "imap-password",
						Usage: "The imap server password",
					},
					&cli.StringFlag{
						Name:  "imap-interval",
						Usage: "The number of seconds to wait before checking for new messages",
						Value: "600", //10 minutes
					},

					&cli.StringFlag{
						Name:  "api-endpoint",
						Usage: "The api server endpoint",
						Value: "http://webapp:3000",
					},
					&cli.StringFlag{
						Name:  "bucket",
						Usage: "The name of the bucket",
					},

					&cli.StringFlag{
						Name:  "amqp-url",
						Usage: "The amqp connection string",
						Value: "amqp://guest:guest@localhost:5672",
					},

					&cli.StringFlag{
						Name:  "amqp-exchange",
						Usage: "The amqp exchange",
						Value: "lodestone",
					},

					&cli.StringFlag{
						Name:  "amqp-queue",
						Usage: "The amqp queue",
						Value: "storagelogs",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(color.HiRedString("ERROR: %v", err))
	}
}
