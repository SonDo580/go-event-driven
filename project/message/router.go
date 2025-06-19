package message

import (
	"encoding/json"
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
	logger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(logger)

	handler := event.NewHandler(receiptsService, spreadsheetsAPI)

	issueReceiptSubscriber := NewRedisSubscriber(rdb, logger, GroupIssueReceipt)
	appendToTrackerSubscriber := NewRedisSubscriber(rdb, logger, GroupAppendToTracker)
	cancelTicketSubscriber := NewRedisSubscriber(rdb, logger, GroupCancelTicket)

	useMiddlewares(router)

	router.AddNoPublisherHandler(
		HandlerIssueReceipt,
		TopicTicketBookingConfirmed,
		issueReceiptSubscriber,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			// Fixing code bug: didn't supply currency for some events => use USD by default
			// TODO: Remove once fixed
			if event.Price.Currency == "" {
				event.Price.Currency = "USD"
			}

			return handler.IssueReceipt(msg.Context(), event)
		},
	)

	router.AddNoPublisherHandler(
		HandlerAppendToTracker,
		TopicTicketBookingConfirmed,
		appendToTrackerSubscriber,
		func(msg *message.Message) error {
			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			// Fixing code bug: didn't supply currency for some events => use USD by default
			// TODO: Remove once fixed
			if event.Price.Currency == "" {
				event.Price.Currency = "USD"
			}

			return handler.AppendToTracker(msg.Context(), event)
		},
	)

	router.AddNoPublisherHandler(
		HandlerCancelTicket,
		TopicTicketBookingCanceled,
		cancelTicketSubscriber,
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
