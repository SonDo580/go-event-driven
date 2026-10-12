package command

import (
	"context"
	"fmt"
	"tickets/entities"
)

func (h Handler) RefundTicket(ctx context.Context, cmd *entities.RefundTicket) error {
	idempotencyKey := cmd.Header.IdempotencyKey
	if idempotencyKey == "" {
		return fmt.Errorf("idempotency key is required")
	}

	return h.receiptsService.VoidReceipt(
		ctx,
		entities.VoidReceipt{
			TicketID:       cmd.TicketID,
			Reason:         "ticket refunded",
			IdempotencyKey: idempotencyKey,
		},
	)
}
