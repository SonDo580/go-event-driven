package message

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, ticketID string) error
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
			ticketID := string(msg.Payload)
			return receiptsService.IssueReceipt(msg.Context(), ticketID)
		},
	)

	router.AddNoPublisherHandler(
		HandlerAppendToTracker,
		TopicAppendToTracker,
		appendToTrackerSubscriber,
		func(msg *message.Message) error {
			ticketID := string(msg.Payload)
			return spreadsheetsAPI.AppendRow(msg.Context(), SheetName, []string{ticketID})
		},
	)

	return router
}
