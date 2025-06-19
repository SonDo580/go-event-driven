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

const brokenMessageID = "2beaf5bc-d5e4-4653-b075-2b36bbf28949"

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
			// Fixing a malformed message
			// TODO: Remove once fixed
			if msg.UUID == brokenMessageID {
				return nil
			}

			if msg.Metadata.Get(constants.MetadataType) != EventTicketBookingConfirmed {
				return nil
			}

			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return handler.IssueReceipt(msg.Context(), event)
		},
	)

	router.AddNoPublisherHandler(
		HandlerAppendToTracker,
		TopicTicketBookingConfirmed,
		appendToTrackerSubscriber,
		func(msg *message.Message) error {
			// Fixing a malformed message
			// TODO: Remove once fixed
			if msg.UUID == brokenMessageID {
				return nil
			}

			if msg.Metadata.Get(constants.MetadataType) != EventTicketBookingConfirmed {
				return nil
			}

			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			return handler.AppendToTracker(msg.Context(), event)
		},
	)

	router.AddNoPublisherHandler(
		HandlerCancelTicket,
		TopicTicketBookingCanceled,
		cancelTicketSubscriber,
		func(msg *message.Message) error {
			if msg.Metadata.Get(constants.MetadataType) != EventTicketBookingCanceled {
				return nil
			}

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
