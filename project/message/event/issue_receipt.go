package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h Handler) IssueReceipt(ctx context.Context, event entities.TicketBookingConfirmed) error {
	slog.Info("Issuing receipt")

	request := entities.IssueReceiptRequest{
		TicketID: event.TicketID,
		Price:    event.Price,
	}

	return h.receiptsService.IssueReceipt(ctx, request)
}
