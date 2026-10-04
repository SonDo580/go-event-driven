package message

import (
	"context"
	"log/slog"
	"tickets/constants"

	"github.com/ThreeDotsLabs/watermill"
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
	watermillLogger watermill.LoggerAdapter,
) {
	issueReceiptSub := NewRedisSubscriber(rdb, watermillLogger, constants.ConsumerGroupIssueReceipt)
	appendToTrackerSub := NewRedisSubscriber(rdb, watermillLogger, constants.ConsumerGroupAppendToTracker)

	go func() {
		messages, err := appendToTrackerSub.Subscribe(context.Background(), constants.TopicAppendToTracker)
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			err := spreadsheetsAPI.AppendRow(
				msg.Context(), "tickets-to-print", []string{string(msg.Payload)},
			)
			if err != nil {
				slog.With("error", err).Error("Error appending to tracker")
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

	go func() {
		messages, err := issueReceiptSub.Subscribe(context.Background(), constants.TopicIssueReceipt)
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			err := receiptsService.IssueReceipt(msg.Context(), string(msg.Payload))
			if err != nil {
				slog.With("error", err).Error("Error issuing receipt")
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()
}
