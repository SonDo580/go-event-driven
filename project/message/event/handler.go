package event

import (
	"context"
	"tickets/entities"
)

type Handler struct {
	receiptsService ReceiptsService
	spreadsheetsAPI SpreadsheetsAPI
}

func NewHandler(
	receiptsService ReceiptsService,
	spreadsheetsAPI SpreadsheetsAPI,
) Handler {
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}

	return Handler{
		receiptsService: receiptsService,
		spreadsheetsAPI: spreadsheetsAPI,
	}
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) error
}
