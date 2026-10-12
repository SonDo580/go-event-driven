package adapters

import (
	"context"
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/payments"
)

type PaymentsServiceClient struct {
	clients *clients.Clients
}

func NewPaymentsServiceClient(clients *clients.Clients) PaymentsServiceClient {
	if clients == nil {
		panic("NewPaymentsServiceClient: client is nil")
	}

	return PaymentsServiceClient{clients: clients}
}

func (c PaymentsServiceClient) RefundPayment(ctx context.Context, request entities.PaymentRefund) error {
	resp, err := c.clients.Payments.PutRefundsWithResponse(
		ctx,
		payments.PaymentRefundRequest{
			PaymentReference: request.TicketID,
			Reason:           request.RefundReason,
			DeduplicationId:  &request.IdempotencyKey,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to put refund for payment %s: %w", request.TicketID, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code for put refund for payment: %d", resp.StatusCode())
	}

	return nil
}
