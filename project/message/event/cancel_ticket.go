package event

import (
	"context"
	"log/slog"
	"tickets/constants"
	"tickets/entities"
)

func (h Handler) CancelTicket(
	ctx context.Context,
	event entities.TicketBookingCanceled,
) error {
	slog.Info("Appending ticket to refund sheet")

	return h.spreadsheetsAPI.AppendRow(
		ctx,
		constants.SheetTicketsToRefund,
		[]string{
			event.TicketID,
			event.CustomerEmail,
			event.Price.Amount,
			event.Price.Currency,
		},
	)
}
