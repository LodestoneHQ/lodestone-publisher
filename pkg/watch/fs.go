package watch

import (
	"fmt"
	"github.com/analogj/fsnotify"
	"github.com/analogj/lodestone-publisher/pkg/model"
	"github.com/analogj/lodestone-publisher/pkg/notify"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

type FsWatcher struct {
	watcher *fsnotify.Watcher
}

func (fs *FsWatcher) Start(notifyClient notify.Interface, config map[string]string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	fs.watcher = watcher
	defer fs.watcher.Close()

	// starting at the root of the specified directory, walk each file/sub-directory searching for
	// new directories
	if err := filepath.Walk(config["dir"], fs.AddWatchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {

			//watch for events
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)

				s3EventName := ""

				if event.Op == fsnotify.Create {
					s3EventName = "s3:ObjectCreated:Put"
				} else if event.Op == fsnotify.CloseWrite {
					s3EventName = "s3:ObjectCreated:Put"
				} else if event.Op == fsnotify.Remove {
					s3EventName = "s3:ObjectRemoved:Delete"
				}

				if s3EventName == "" {
					//ignore event
					break
				}

				s3EventPayload := model.S3Event{}
				err := s3EventPayload.Create("fs", s3EventName, config["bucket"], url.PathEscape(event.Name), event.Name)

				if err == nil {
					err := notifyClient.Publish(s3EventPayload)
					if err != nil {
						fmt.Print(err)
					}

				} else {
					//log an error message if we cant create a valid S3EventPayload. Then ignore.
					fmt.Print(err)
				}

			//watch for errors
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	<-done
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func (fs *FsWatcher) AddWatchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return fs.watcher.Add(path)
	}

	return nil
}
