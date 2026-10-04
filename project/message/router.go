package message

import (
	"context"
	"encoding/json"
	"log/slog"
	"tickets/constants"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) error
}

func NewWatermillRouter(
	receiptsService ReceiptsService,
	spreadsheetsAPI SpreadsheetsAPI,
	rdb *redis.Client,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)

	issueReceiptSub := NewRedisSubscriber(rdb, watermillLogger, constants.ConsumerGroupIssueReceipt)
	appendToTrackerSub := NewRedisSubscriber(rdb, watermillLogger, constants.ConsumerGroupAppendToTracker)

	router.AddConsumerHandler(
		constants.HandlerIssueReceipt,
		constants.TopicIssueReceipt,
		issueReceiptSub,
		func(msg *message.Message) error {
			var payload entities.IssueReceiptPayload
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}

			slog.Info("Issuing receipt")

			request := entities.IssueReceiptRequest{
				TicketID: payload.TicketID,
				Price:    payload.Price,
			}

			return receiptsService.IssueReceipt(msg.Context(), request)
		},
	)

	router.AddConsumerHandler(
		constants.HandlerAppendToTracker,
		constants.TopicAppendToTracker,
		appendToTrackerSub,
		func(msg *message.Message) error {
			var payload entities.AppendToTrackerPayload
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}

			slog.Info("Appending ticket to the tracker")

			return spreadsheetsAPI.AppendRow(
				msg.Context(),
				"tickets-to-print",
				[]string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency},
			)
		},
	)

	return router
}
