package command

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
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
	eventBus        *cqrs.EventBus
}

func NewHandler(
	receiptsService ReceiptsService,
	paymentsService PaymentsService,
	eventBus *cqrs.EventBus,
) Handler {
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if paymentsService == nil {
		panic("missing paymentsService")
	}
	if eventBus == nil {
		panic("missing eventBus")
	}

	return Handler{
		receiptsService: receiptsService,
		paymentsService: paymentsService,
		eventBus:        eventBus,
	}
}
