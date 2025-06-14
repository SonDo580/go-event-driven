package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"

	"tickets/constants"
	"tickets/entities"
	ticketsMsg "tickets/message"
)

type ticketsStatusRequest struct {
	Tickets []ticketStatusRequest `json:"tickets"`
}

type ticketStatusRequest struct {
	TicketID      string         `json:"ticket_id"`
	Status        string         `json:"status"`
	Price         entities.Money `json:"price"`
	CustomerEmail string         `json:"customer_email"`
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	var request ticketsStatusRequest
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

			payload, err := json.Marshal(event)
			if err != nil {
				return err
			}

			err = h.publisher.Publish(
				ticketsMsg.TopicTicketBookingConfirmed,
				message.NewMessage(watermill.NewUUID(), []byte(payload)),
			)
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

			payload, err := json.Marshal(event)
			if err != nil {
				return err
			}

			err = h.publisher.Publish(
				ticketsMsg.TopicTicketBookingCanceled,
				message.NewMessage(watermill.NewUUID(), []byte(payload)),
			)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}

	}

	return c.NoContent(http.StatusOK)
}
