package message

import (
	"tickets/message/event"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewWatermillRouter(
	eventHandler event.Handler,
	rdb *redis.Client,
	logger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(logger)
	useMiddlewares(router)

	eventProcessor := NewEventProcessor(rdb, router, logger)
	RegisterEventHandlers(
		eventProcessor,
		[]cqrs.EventHandler{
			cqrs.NewEventHandler(
				"IssueReceipt",
				eventHandler.IssueReceipt,
			),
			cqrs.NewEventHandler(
				"AppendToTracker",
				eventHandler.AppendToTracker,
			),
			cqrs.NewEventHandler(
				"CancelTicket",
				eventHandler.CancelTicket,
			),
			cqrs.NewEventHandler(
				"StoreTicket",
				eventHandler.StoreTicket,
			),
		},
	)

	return router
}
