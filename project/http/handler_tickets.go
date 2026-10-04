package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tickets/constants"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
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
	var request TicketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		if ticket.Status != constants.TicketStatusConfirmed {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}

		issueReceiptPayload := entities.IssueReceiptPayload{
			TicketID: ticket.TicketID,
			Price:    ticket.Price,
		}
		issueReceiptJSON, err := json.Marshal(issueReceiptPayload)
		if err != nil {
			return err
		}

		msg := message.NewMessage(watermill.NewUUID(), issueReceiptJSON)
		err = h.publisher.Publish(constants.TopicIssueReceipt, msg)
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

		msg = message.NewMessage(watermill.NewUUID(), appendToTrackerJSON)
		err = h.publisher.Publish(constants.TopicAppendToTracker, msg)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
