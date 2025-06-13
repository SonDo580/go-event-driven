package event

import (
	"context"
	"log/slog"
	"tickets/constants"
	"tickets/entities"
)

func (h Handler) AppendToTracker(
	ctx context.Context,
	event entities.TicketBookingConfirmed,
) error {
	slog.Info("Appending ticket to tracker")
	return h.spreadsheetsAPI.AppendRow(
		ctx,
		constants.SheetName,
		[]string{
			event.TicketID,
			event.CustomerEmail,
			event.Price.Amount,
			event.Price.Currency,
		},
	)
}
