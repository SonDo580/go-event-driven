package event

import (
	"context"
	"tickets/constants"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

func (h Handler) AppendToTracker(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Appending ticket to tracker")

	return h.spreadsheetsAPI.AppendRow(
		ctx,
		constants.SheetTicketsToPrint,
		[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency},
	)
}

func (h Handler) NewAppendToTrackerHandler() cqrs.EventHandler {
	return cqrs.NewEventHandler(
		constants.HandlerAppendToTracker,
		h.AppendToTracker,
	)
}
