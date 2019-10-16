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

	//b, err := json.Marshal(event)
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	//
	//var inInterface map[string]interface{}
	//json.Unmarshal(b, &inInterface)

	fmt.Print(n.client.HSet(n.queue, "testfieldname", event))
	return nil
	//
	////n.client.RPush()
	//err := n.client.Set("key", "value", 0).Err()
	//if err != nil {
	//	panic(err)
	//}
	//
	//val, err := n.client.Get("key").Result()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("key", val)
	//
	//val2, err := n.client.Get("key2").Result()
	//if err == redis.Nil {
	//	fmt.Println("key2 does not exist")
	//} else if err != nil {
	//	panic(err)
	//} else {
	//	fmt.Println("key2", val2)
	//}
	//// Output: key value
	//// key2 does not exist
	//return nil
}
