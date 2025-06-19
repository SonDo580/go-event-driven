package main

import (
	"context"
	"sync"
)

type IssueReceiptRequest struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) error
}

type ReceiptsServiceStub struct {
	lock           sync.Mutex
	IssuedReceipts []IssueReceiptRequest
}

func (stub *ReceiptsServiceStub) IssueReceipt(
	ctx context.Context, request IssueReceiptRequest,
) error {
	stub.lock.Lock()
	defer stub.lock.Unlock()

	stub.IssuedReceipts = append(stub.IssuedReceipts, request)
	return nil
}
