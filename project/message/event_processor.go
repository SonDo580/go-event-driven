package message

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewEventProcessor(
	rdb *redis.Client,
	router *message.Router,
	logger watermill.LoggerAdapter,
) *cqrs.EventProcessor {
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				consumerGroup := "svc-ticket." + params.HandlerName
				return NewRedisSubscriber(rdb, logger, consumerGroup), nil
			},
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
			Logger: logger,
		},
	)

	if err != nil {
		panic(err)
	}

	return eventProcessor
}

func RegisterEventHandlers(
	eventProcessor *cqrs.EventProcessor,
	handlers []cqrs.EventHandler,
) {
	err := eventProcessor.AddHandlers(handlers...)
	if err != nil {
		panic(err)
	}
}
