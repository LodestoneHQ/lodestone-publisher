package notify

import "github.com/analogj/lodestone-publisher/pkg/model"

type Interface interface {
	Init(config map[string]string) error
	Publish(event model.S3Event) error
}
