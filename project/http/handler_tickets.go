package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"tickets/constants"
	"tickets/entities"
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
	var request TicketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		if ticket.Status == constants.TicketStatusConfirmed {
			event := entities.TicketBookingConfirmed{
				Header:        entities.NewEventHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			err := h.eventBus.Publish(c.Request().Context(), event)
			if err != nil {
				return err
			}
		} else if ticket.Status == constants.TicketStatusCanceled {
			event := entities.TicketBookingCanceled{
				Header:        entities.NewEventHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			err := h.eventBus.Publish(c.Request().Context(), event)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}

	}

	return c.NoContent(http.StatusOK)
}

func (h Handler) GetTickets(c echo.Context) error {
	tickets, err := h.ticketsRepo.FindAll(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to retrieve tickets: %w", err)
	}
	return c.JSON(http.StatusOK, tickets)
}
