package message

import (
	"context"
	"tickets/constants"

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
			return receiptsService.IssueReceipt(msg.Context(), string(msg.Payload))
		},
	)

	router.AddConsumerHandler(
		constants.HandlerAppendToTracker,
		constants.TopicAppendToTracker,
		appendToTrackerSub,
		func(msg *message.Message) error {
			return spreadsheetsAPI.AppendRow(
				msg.Context(), "tickets-to-print", []string{string(msg.Payload)},
			)
		},
	)

	return router
}
