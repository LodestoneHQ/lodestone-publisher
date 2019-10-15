package notify

import "github.com/analogj/lodestone-fs-watcher/pkg/model"

type Interface interface {
	Init(config map[string]string) error
	Publish(event model.S3Event) error
}
