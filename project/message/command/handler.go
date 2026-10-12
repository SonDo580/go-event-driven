package command

import (
	"context"
	"tickets/entities"
)

type ReceiptsService interface {
	VoidReceipt(ctx context.Context, command entities.VoidReceipt) error
}

type Handler struct {
	receiptsService ReceiptsService
}

func NewHandler(
	receiptsService ReceiptsService,
) Handler {
	if receiptsService == nil {
		panic("missing receiptsService")
	}

	return Handler{
		receiptsService: receiptsService,
	}
}
