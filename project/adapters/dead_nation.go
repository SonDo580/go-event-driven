package adapters

import (
	"context"
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
	"github.com/google/uuid"
)

type DeadNationClient struct {
	clients *clients.Clients
}

func NewDeadNationClient(clients *clients.Clients) *DeadNationClient {
	if clients == nil {
		panic("NewDeadNationClient: clients is nil")
	}

	return &DeadNationClient{clients: clients}
}

func (c DeadNationClient) Book(ctx context.Context, booking entities.DeadNationBooking) error {
	bookingID, err := uuid.Parse(booking.BookingID)
	if err != nil {
		return err
	}

	deadNationID, err := uuid.Parse(booking.DeadNationEventID)
	if err != nil {
		return err
	}

	resp, err := c.clients.DeadNation.PostTicketBookingWithResponse(
		ctx, dead_nation.PostTicketBookingRequest{
			BookingId:       bookingID,
			EventId:         deadNationID,
			NumberOfTickets: booking.NumberOfTickets,
			CustomerAddress: booking.CustomerEmail,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to book in Dead Nation: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code from Dead Nation: %d", resp.StatusCode())
	}

	return nil
}
