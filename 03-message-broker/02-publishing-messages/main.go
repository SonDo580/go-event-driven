package main

import (
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

const topic = "progress"

func main() {
	logger := watermill.NewSlogLogger(nil)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)

	if err != nil {
		panic(err)
	}

	for _, payload := range []string{"50", "100"} {
		msg := message.NewMessage(watermill.NewUUID(), []byte(payload))
		err = publisher.Publish(topic, msg)
		if err != nil {
			panic(err)
		}
	}
}
