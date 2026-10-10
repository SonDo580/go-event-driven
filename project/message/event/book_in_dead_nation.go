package event

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) BookInDeadNation(ctx context.Context, event *entities.BookingMade) error {
	log.FromContext(ctx).Info("Booking tickets in Dead Nation")

	show, err := h.showsRepository.ShowByID(ctx, event.ShowID)
	if err != nil {
		return err
	}

	request := entities.DeadNationBooking{
		BookingID:         event.BookingID,
		DeadNationEventID: show.DeadNationID,
		NumberOfTickets:   event.NumberOfTickets,
		CustomerEmail:     event.CustomerEmail,
	}

	err = h.deadNationAPI.Book(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to book in Dead Nation: %w", err)
	}

	return nil
}
