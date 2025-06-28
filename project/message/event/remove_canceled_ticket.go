package event

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) RemoveCanceledTicket(
	ctx context.Context,
	event *entities.TicketBookingCanceled,
) error {
	log.FromContext(ctx).Info("Removing ticket")
	return h.ticketsRepository.Remove(ctx, event.TicketID)
}
