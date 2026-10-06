package http

import (
	"fmt"
	"net/http"
	"tickets/constants"
	"tickets/entities"

	"github.com/labstack/echo/v4"
)

type TicketsStatusRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

type TicketStatusRequest struct {
	TicketID      string         `json:"ticket_id"`
	Status        string         `json:"status"`
	Price         entities.Money `json:"price"`
	CustomerEmail string         `json:"customer_email"`
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	idempotencyKey := c.Request().Header.Get(constants.HeaderIdempotencyKey)
	if idempotencyKey == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	var request TicketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		switch ticket.Status {
		case constants.TicketStatusConfirmed:
			event := entities.TicketBookingConfirmed{
				Header:        entities.NewMessageHeaderWithIdempotencyKey(idempotencyKey + ticket.TicketID),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			if err = h.eventBus.Publish(c.Request().Context(), event); err != nil {
				return fmt.Errorf("failed to publish TicketBookingConfirmed event: %w", err)
			}
		case constants.TicketStatusCanceled:
			event := entities.TicketBookingCanceled{
				Header:        entities.NewMessageHeaderWithIdempotencyKey(idempotencyKey + ticket.TicketID),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			if err = h.eventBus.Publish(c.Request().Context(), event); err != nil {
				return fmt.Errorf("failed to publish TicketBookingCanceled event: %w", err)
			}
		default:
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h Handler) GetTickets(c echo.Context) error {
	tickets, err := h.ticketsRepo.FindAll(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to find tickets: %w", err)
	}

	return c.JSON(http.StatusOK, tickets)
}
