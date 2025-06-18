package event

import (
	"context"
	"tickets/constants"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) AppendToTracker(
	ctx context.Context,
	event entities.TicketBookingConfirmed,
) error {
	log.FromContext(ctx).Info("Appending ticket to tracker")

	return h.spreadsheetsAPI.AppendRow(
		ctx,
		constants.SheetTicketsToPrint,
		[]string{
			event.TicketID,
			event.CustomerEmail,
			event.Price.Amount,
			event.Price.Currency,
		},
	)
}
