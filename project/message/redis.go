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
		Addr:     addr,
		PoolSize: 200,
	})
}

func NewRedisPublisher(rdb *redis.Client, watermillLogger watermill.LoggerAdapter) message.Publisher {
	var pub message.Publisher

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	// - Pass the request context when publishing message.
	//   The decorator will read the request context for correlation ID
	//   and add 'correlation_id' metadata to messages.
	// - Prerequisite: Set correlation ID in request context
	//   (extract from 'Correlation-ID' request header).
	pub = log.CorrelationPublisherDecorator{Publisher: pub}

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
