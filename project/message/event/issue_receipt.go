package event

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) IssueReceipt(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Issuing receipt")

	request := entities.IssueReceiptRequest{
		IdempotencyKey: event.Header.IdempotencyKey,
		TicketID:       event.TicketID,
		Price:          event.Price,
	}

	return h.receiptsService.IssueReceipt(ctx, request)
}
