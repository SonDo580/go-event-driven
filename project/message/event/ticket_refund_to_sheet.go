package event

import (
	"context"
	"tickets/constants"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) CancelTicket(ctx context.Context, event entities.TicketBookingCanceled) error {
	log.FromContext(ctx).Info("Adding ticket refund to sheet")

	return h.spreadsheetsAPI.AppendRow(
		ctx,
		constants.SheetTicketsToRefund,
		[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency},
	)
}
