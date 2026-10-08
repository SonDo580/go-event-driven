package http

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type TicketsRepository interface {
	FindAll(ctx context.Context) ([]entities.Ticket, error)
}

type ShowsRepository interface {
	Add(ctx context.Context, show entities.Show) error
}

type BookingsRepository interface {
	Add(ctx context.Context, booking entities.Booking) error
}

type Handler struct {
	eventBus     *cqrs.EventBus
	ticketsRepo  TicketsRepository
	showsRepo    ShowsRepository
	bookingsRepo BookingsRepository
}
