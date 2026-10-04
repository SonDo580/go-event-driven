package message

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		PoolSize: 200,
	})
}

func NewRedisPublisher(rdb *redis.Client, watermillLogger watermill.LoggerAdapter) message.Publisher {
	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	return pub
}

func NewRedisSubscriber(
	rdb *redis.Client, watermillLogger watermill.LoggerAdapter, consumerGroup string,
) message.Subscriber {
	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: consumerGroup,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	return sub
}
