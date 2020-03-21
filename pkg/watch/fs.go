package watch

import (
	"fmt"
	"github.com/analogj/fsnotify"
	"github.com/analogj/lodestone-publisher/pkg/model"
	"github.com/analogj/lodestone-publisher/pkg/notify"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type FsWatcher struct {
	logger  *logrus.Entry
	watcher *fsnotify.Watcher
}

func (fs *FsWatcher) Start(logger *logrus.Entry, notifyClient notify.Interface, config map[string]string) {
	fs.logger = logger
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fs.logger.Fatal(err)
	}
	fs.watcher = watcher
	defer fs.watcher.Close()

	// starting at the root of the specified directory, walk each file/sub-directory searching for
	// new directories
	if err := filepath.Walk(config["dir"], fs.AddWatchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	fmt.Println("Start watching for filesystem events")
	done := make(chan bool)

	go func() {
		for {
			select {

			//watch for events
			case event, ok := <-watcher.Events:
				if !ok {
					fs.logger.Warnln("FAILED event:", event)
					return
				}
				fs.logger.Debugln("event:", event)

				// PSEUDO CODE
				// check if event is "add" or "delete"
				// if event is "add" and is a file:
				// 	 generate an event and publish
				// if event is "add" and is a folder:
				// 	 add watcher
				// if event is "remove" and is a file or folder:
				//   generate an event and publish
				//   remove watcher (in-case this is a folder)

				s3EventName := ""
				if (event.Op&fsnotify.Create == fsnotify.Create) || (event.Op&fsnotify.CloseWrite == fsnotify.CloseWrite) {
					fs.logger.Infoln("Processing create event: ", event)

					s3EventName = "s3:ObjectCreated:Put"

					//get event file/folder data.
					eventPathInfo, err := os.Stat(event.Name)
					if fs.CheckErr(err) {
						break
					}

					switch mode := eventPathInfo.Mode(); {
					case mode.IsDir():
						// newly added folder
						err := fs.AddWatchDir(event.Name, eventPathInfo, nil)
						fs.CheckErr(err)

					case mode.IsRegular():
						// newly added file.
						s3Event, err := GenerateS3Event(s3EventName, event, config)
						fs.CheckErr(err)
						if err == nil {
							err := notifyClient.Publish(s3Event)
							fs.CheckErr(err)
						}
					}

				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					fs.logger.Infoln("Processing delete event: ", event)

					s3EventName = "s3:ObjectRemoved:Delete"

					s3Event, err := GenerateS3Event(s3EventName, event, config)
					fs.CheckErr(err)
					if err == nil {
						err := notifyClient.Publish(s3Event)
						fs.CheckErr(err)
					}

					fs.RemoveWatchDir(event.Name, nil, nil)
				} else {
					fs.logger.Infoln("Ignoring event: ", event)
				}

			//watch for errors
			case err, ok := <-watcher.Errors:
				if !ok {
					fs.logger.Errorln("failed error", err)
					return
				}
				fs.logger.Errorln("error:", err)
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
		fs.logger.Infof("Watching new directory: %v", path)
		return fs.watcher.Add(path)
	}

	return nil
}

func (fs *FsWatcher) RemoveWatchDir(path string, fi os.FileInfo, err error) error {
	fs.logger.Infof("Removing watch directory: %v", path)
	return fs.watcher.Remove(path)
}

// Helpers

func GenerateS3Event(s3EventName string, fsevent fsnotify.Event, config map[string]string) (model.S3Event, error) {

	relPath, err := filepath.Rel(config["dir"], fsevent.Name)
	if err != nil {
		return model.S3Event{}, err
	}

	s3EventPayload := model.S3Event{}
	err = s3EventPayload.Create("fs", s3EventName, config["bucket"], relPath, fsevent.Name)
	return s3EventPayload, err
}

func (fs *FsWatcher) CheckErr(err error) bool {
	if err != nil {
		fs.logger.Errorln("error:", err)
		return true
	} else {
		return false
	}
}
