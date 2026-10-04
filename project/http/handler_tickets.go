package http

import (
	"net/http"
	"tickets/constants"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request ticketsConfirmationRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		payload := []byte(ticket)

		msg := message.NewMessage(watermill.NewUUID(), payload)
		err = h.publisher.Publish(constants.TopicIssueReceipt, msg)
		if err != nil {
			return err
		}

		msg = message.NewMessage(watermill.NewUUID(), payload)
		err = h.publisher.Publish(constants.TopicAppendToTracker, msg)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
