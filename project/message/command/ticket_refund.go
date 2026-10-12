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

	err := h.receiptsService.VoidReceipt(
		ctx,
		entities.VoidReceipt{
			TicketID:       cmd.TicketID,
			Reason:         "ticket refunded",
			IdempotencyKey: idempotencyKey,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to void receipt: %w", err)
	}

	err = h.paymentsService.RefundPayment(
		ctx,
		entities.PaymentRefund{
			TicketID:       cmd.TicketID,
			RefundReason:   "ticket refunded",
			IdempotencyKey: idempotencyKey,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to refund payment: %w", err)
	}

	return nil
}
