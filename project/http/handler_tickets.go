package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"

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
		if ticket.Status != "confirmed" {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}

		err = h.publisher.Publish(
			ticketsMsg.TopicIssueReceipt,
			message.NewMessage(watermill.NewUUID(), []byte(ticket.TicketID)),
		)
		if err != nil {
			return err
		}

		appendToTrackerPayload := entities.AppendToTrackerPayload{
			TicketID:      ticket.TicketID,
			CustomerEmail: ticket.CustomerEmail,
			Price:         ticket.Price,
		}

		appendToTrackerJSON, err := json.Marshal(appendToTrackerPayload)
		if err != nil {
			return err
		}

		err = h.publisher.Publish(
			ticketsMsg.TopicAppendToTracker,
			message.NewMessage(watermill.NewUUID(), []byte(appendToTrackerJSON)),
		)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
