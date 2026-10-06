package event

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

func (h Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Printing ticket")

	fileName := event.TicketID + "-ticket.html"
	fileContent := `
		<html>
			<head>
				<title>Ticket</title>
			</head>
			<body>
				<h1>Ticket ` + event.TicketID + `</h1>
				<p>Price: ` + event.Price.Amount + ` ` + event.Price.Currency + `</p>	
			</body>
		</html>
	`

	err := h.filesAPI.UploadFile(ctx, fileName, fileContent)
	if err != nil {
		return err
	}

	newEvent := entities.TicketPrinted{
		Header:   entities.NewMessageHeader(),
		TicketID: event.TicketID,
		FileName: fileName,
	}
	if err = h.eventBus.Publish(ctx, newEvent); err != nil {
		return fmt.Errorf("failed to publish TicketPrinted event: %w", err)
	}

	return nil
}
