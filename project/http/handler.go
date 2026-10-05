package http

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type TicketsRepository interface {
	FindAll(ctx context.Context) ([]entities.Ticket, error)
}

type Handler struct {
	eventBus    *cqrs.EventBus
	ticketsRepo TicketsRepository
}
