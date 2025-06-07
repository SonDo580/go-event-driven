package http

import (
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"

	ticketsMsg "tickets/message"
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
		err = h.publisher.Publish(
			ticketsMsg.TopicIssueReceipt,
			message.NewMessage(watermill.NewUUID(), []byte(ticket)),
		)
		if err != nil {
			return err
		}

		err = h.publisher.Publish(
			ticketsMsg.TopicAppendToTracker,
			message.NewMessage(watermill.NewUUID(), []byte(ticket)),
		)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
