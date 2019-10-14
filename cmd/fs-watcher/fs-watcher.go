package main

import (
	"fmt"
	"github.com/analogj/lodestone-fs-watcher/pkg/version"
	"os"
	"time"

	"github.com/analogj/lodestone-fs-watcher/pkg"
	"github.com/urfave/cli"
	"path/filepath"

	"github.com/analogj/go-util/utils"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "lodestone-fs-watcher",
		Usage:    "Filesystem watcher for lodestone",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			capsuleUrl := "https://github.com/AnalogJ/lodestone-fs-watcher"

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
				Usage: "Start the Lodestone filesystem watcher",
				Action: func(c *cli.Context) error {

					return nil
				},

				Flags: []cli.Flag{},
			},
		},
	}

	app.Run(os.Args)
}
