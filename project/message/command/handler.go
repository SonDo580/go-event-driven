package command

import (
	"context"
	"tickets/entities"
)

type ReceiptsService interface {
	VoidReceipt(ctx context.Context, command entities.VoidReceipt) error
}

type PaymentsService interface {
	RefundPayment(ctx context.Context, request entities.PaymentRefund) error
}

type Handler struct {
	receiptsService ReceiptsService
	paymentsService PaymentsService
}

func NewHandler(
	receiptsService ReceiptsService,
	paymentsService PaymentsService,
) Handler {
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if paymentsService == nil {
		panic("missing paymentsService")
	}

	return Handler{
		receiptsService: receiptsService,
		paymentsService: paymentsService,
	}
}
