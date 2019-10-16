package notify

import (
	"fmt"
	"github.com/analogj/lodestone-publisher/pkg/model"
)
import "github.com/go-redis/redis"

type RedisNotify struct {
	client *redis.Client
	queue  string
}

func (n *RedisNotify) Init(config map[string]string) error {
	n.client = redis.NewClient(&redis.Options{
		Addr:     config["addr"],     //"localhost:6379",
		Password: config["password"], //"", // no password set
		DB:       0,                  // use default DB
	})
	n.queue = config["queue"]

	pong, err := n.client.Ping().Result()
	fmt.Println(pong, err)

	return err
}

func (n *RedisNotify) Publish(event model.S3Event) error {
	fmt.Println("Publishing event..")

	resp := n.client.HSet(n.queue, fmt.Sprintf("%s/%s", event.Records[0].S3.Bucket.Name, event.Records[0].S3.Object.Key), event)
	return resp.Err()
}
