package message

import (
	"encoding/json"
	"tickets/constants"
	"tickets/entities"
	"tickets/message/event"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewWatermillRouter(
	receiptsService event.ReceiptsService,
	spreadsheetsAPI event.SpreadsheetsAPI,
	rdb *redis.Client,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)
	handler := event.NewHandler(spreadsheetsAPI, receiptsService)
	useMiddlewares(router, watermillLogger)

	issueReceiptSub := NewRedisSubscriber(rdb, watermillLogger, constants.ConsumerGroupIssueReceipt)
	appendToTrackerSub := NewRedisSubscriber(rdb, watermillLogger, constants.ConsumerGroupAppendToTracker)
	cancelTicketSub := NewRedisSubscriber(rdb, watermillLogger, constants.ConsumerGroupRefund)

	router.AddConsumerHandler(
		constants.HandlerIssueReceipt,
		constants.TopicTicketBookingConfirmed,
		issueReceiptSub,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			// TODO: remove once fix
			if event.Price.Currency == "" {
				event.Price.Currency = "USD"
			}

			return handler.IssueReceipt(msg.Context(), event)
		},
	)

	router.AddConsumerHandler(
		constants.HandlerAppendToTracker,
		constants.TopicTicketBookingConfirmed,
		appendToTrackerSub,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			// TODO: remove once fix
			if event.Price.Currency == "" {
				event.Price.Currency = "USD"
			}

			return handler.AppendToTracker(msg.Context(), event)
		},
	)

	router.AddConsumerHandler(
		constants.HandlerCancelTicket,
		constants.TopicTicketBookingCanceled,
		cancelTicketSub,
		func(msg *message.Message) error {
			var event entities.TicketBookingCanceled
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return handler.CancelTicket(msg.Context(), event)
		},
	)

	return router
}
