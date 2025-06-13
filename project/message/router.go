package message

import (
	"context"
	"encoding/json"
	"log/slog"
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
	logger watermill.LoggerAdapter,
) *message.Router {
	router := message.NewDefaultRouter(logger)

	issueReceiptSubscriber := NewRedisSubscriber(rdb, logger, GroupIssueReceipt)
	appendToTrackerSubscriber := NewRedisSubscriber(rdb, logger, GroupAppendToTracker)

	router.AddNoPublisherHandler(
		HandlerIssueReceipt,
		TopicIssueReceipt,
		issueReceiptSubscriber,
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

	router.AddNoPublisherHandler(
		HandlerAppendToTracker,
		TopicAppendToTracker,
		appendToTrackerSubscriber,
		func(msg *message.Message) error {
			var payload entities.AppendToTrackerPayload
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}

			slog.Info("Appending ticket to tracker")

			return spreadsheetsAPI.AppendRow(
				msg.Context(),
				SheetName,
				[]string{
					payload.TicketID,
					payload.CustomerEmail,
					payload.Price.Amount,
					payload.Price.Currency,
				},
			)
		},
	)

	return router
}
