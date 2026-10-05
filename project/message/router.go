package message

import (
	"tickets/constants"
	"tickets/message/event"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewWatermillRouter(
	eventHandler event.Handler,
	eventProcessConfig cqrs.EventProcessorConfig,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)
	useMiddlewares(router, watermillLogger)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessConfig)
	if err != nil {
		panic(err)
	}

	eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			constants.HandlerAppendToTracker,
			eventHandler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			constants.HandlerIssueReceipt,
			eventHandler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			constants.HandlerCancelTicket,
			eventHandler.CancelTicket,
		),
		cqrs.NewEventHandler(
			constants.HandlerStoreTicket,
			eventHandler.StoreTicket,
		),
	)

	return router
}
