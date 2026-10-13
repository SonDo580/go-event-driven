package message

import (
	"tickets/db"
	"tickets/message/command"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewWatermillRouter(
	postgresSub message.Subscriber,
	publisher message.Publisher,
	eventProcessorConfig cqrs.EventProcessorConfig,
	eventHandler event.Handler,
	commandProcessorConfig cqrs.CommandProcessorConfig,
	commandHandler command.Handler,
	opsBookingReadModel db.OpsBookingReadModel,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)

	useMiddlewares(router, watermillLogger)

	outbox.AddForwarderHandler(postgresSub, publisher, router, watermillLogger)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		panic(err)
	}

	eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"BookInDeadNation",
			eventHandler.BookInDeadNation,
		),
		cqrs.NewEventHandler(
			"AppendToTracker",
			eventHandler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			"IssueReceipt",
			eventHandler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"PrintTicket",
			eventHandler.PrintTicket,
		),
		cqrs.NewEventHandler(
			"StoreTicket",
			eventHandler.StoreTicket,
		),
		cqrs.NewEventHandler(
			"CancelTicket",
			eventHandler.CancelTicket,
		),
		cqrs.NewEventHandler(
			"RemoveCanceledTicket",
			eventHandler.RemoveCanceledTicket,
		),

		cqrs.NewEventHandler(
			"opsBookingReadModel.OnBookingMade",
			opsBookingReadModel.OnBookingMade,
		),
		cqrs.NewEventHandler(
			"opsBookingReadModel.OnTicketBookingConfirmed",
			opsBookingReadModel.OnTicketBookingConfirmed,
		),
		cqrs.NewEventHandler(
			"opsBookingReadModel.OnTicketReceiptIssued",
			opsBookingReadModel.OnTicketReceiptIssued,
		),
		cqrs.NewEventHandler(
			"opsBookingReadModel.OnTicketPrinted",
			opsBookingReadModel.OnTicketPrinted,
		),
		cqrs.NewEventHandler(
			"opsBookingReadModel.OnTicketRefunded",
			opsBookingReadModel.OnTicketRefunded,
		),
	)

	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(router, commandProcessorConfig)
	if err != nil {
		panic(err)
	}

	commandProcessor.AddHandler(
		cqrs.NewCommandHandler(
			"RefundTicket",
			commandHandler.RefundTicket,
		),
	)

	return router
}
