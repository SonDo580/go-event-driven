package event

import (
	"context"
	"tickets/entities"
)

type Handler struct {
	receiptsService   ReceiptsService
	spreadsheetsAPI   SpreadsheetsAPI
	ticketsRepository TicketsRepository
}

func NewHandler(
	receiptsService ReceiptsService,
	spreadsheetsAPI SpreadsheetsAPI,
	ticketsRepository TicketsRepository,
) Handler {
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if ticketsRepository == nil {
		panic("missing ticketsRepository")
	}

	return Handler{
		receiptsService:   receiptsService,
		spreadsheetsAPI:   spreadsheetsAPI,
		ticketsRepository: ticketsRepository,
	}
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) error
}

type TicketsRepository interface {
	Add(ctx context.Context, ticket entities.Ticket) error
}
