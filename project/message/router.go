package message

import (
	"tickets/constants"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewWatermillRouter(
	postgresSub message.Subscriber,
	publisher message.Publisher,
	eventProcessConfig cqrs.EventProcessorConfig,
	eventHandler event.Handler,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)

	useMiddlewares(router, watermillLogger)

	outbox.AddForwarderHandler(postgresSub, publisher, router, watermillLogger)

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
			constants.HandlerPrintTicket,
			eventHandler.PrintTicket,
		),
		cqrs.NewEventHandler(
			constants.HandlerStoreTicket,
			eventHandler.StoreTicket,
		),
		cqrs.NewEventHandler(
			constants.HandlerCancelTicket,
			eventHandler.CancelTicket,
		),
		cqrs.NewEventHandler(
			constants.HandlerRemoveCanceledTicket,
			eventHandler.RemoveCanceledTicket,
		),
	)

	return router
}
