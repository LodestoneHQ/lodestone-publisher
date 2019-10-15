package watch

import (
	"fmt"
	"github.com/analogj/fsnotify"
	"github.com/analogj/lodestone-fs-watcher/pkg/model"
	"github.com/analogj/lodestone-fs-watcher/pkg/notify"
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
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)

					var s3EventName string

					switch event.Op {
					case fsnotify.CloseWrite:
						s3EventName = "s3:ObjectCreated:Put"
					case fsnotify.Remove:
						s3EventName = "s3:ObjectRemoved:Delete"
					}

					s3EventPayload := model.S3Event{}
					event, err := s3EventPayload.Create("fs", s3EventName, config["bucket"], url.PathEscape(event.Name), event.Name)

					if err == nil {
						err := notifyClient.Publish(event)
						if err != nil {
							fmt.Print(err)
						}

					} else {
						//log an error message if we cant create a valid S3EventPayload. Then ignore.
						fmt.Print(err)
					}

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
