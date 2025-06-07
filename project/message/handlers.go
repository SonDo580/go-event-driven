package message

import (
	"context"
	"log/slog"

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

func NewHandlers(
	receiptsService ReceiptsService,
	spreadsheetsAPI SpreadsheetsAPI,
	rdb *redis.Client,
	logger watermill.LoggerAdapter,
) []func() {
	issueReceiptSubscriber := NewRedisSubscriber(rdb, logger, GroupIssueReceipt)
	appendToTrackerSubscriber := NewRedisSubscriber(rdb, logger, GroupAppendToTracker)

	return []func(){
		func() {
			processMessages(
				issueReceiptSubscriber,
				TopicIssueReceipt,
				receiptsService.IssueReceipt,
				"Error issuing receipt",
			)
		},
		func() {
			processMessages(
				appendToTrackerSubscriber,
				TopicAppendToTracker,
				func(ctx context.Context, ticketID string) error {
					return spreadsheetsAPI.AppendRow(ctx, SheetName, []string{ticketID})
				},
				"Error appending to tracker",
			)
		},
	}
}

func processMessages(
	subscriber message.Subscriber,
	topic string,
	action func(ctx context.Context, ticketID string) error,
	actionFailedMsg string,
) {
	messages, err := subscriber.Subscribe(context.Background(), topic)
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		ticketID := string(msg.Payload)
		err := action(msg.Context(), ticketID)

		if err != nil {
			slog.With("error", err).Error(actionFailedMsg)
			msg.Nack()
		} else {
			msg.Ack()
		}
	}
}
