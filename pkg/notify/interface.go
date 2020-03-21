package notify

import (
	"github.com/analogj/lodestone-publisher/pkg/model"
	"github.com/sirupsen/logrus"
)

type Interface interface {
	Init(logger *logrus.Entry, config map[string]string) error
	Publish(event model.S3Event) error
	Close() error
}
