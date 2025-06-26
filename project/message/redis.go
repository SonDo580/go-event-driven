package message

import (
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func NewRedisPublisher(rdb *redis.Client, logger watermill.LoggerAdapter) message.Publisher {
	var publisher message.Publisher
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)

	if err != nil {
		panic(err)
	}

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}

	return publisher
}

func NewRedisSubscriber(rdb *redis.Client, logger watermill.LoggerAdapter, consumerGroup string) message.Subscriber {
	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: consumerGroup,
	}, logger)

	if err != nil {
		panic(err)
	}
	return subscriber
}
