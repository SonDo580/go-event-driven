package adapters

import (
	"context"
	"sync"
	"tickets/entities"
)

type ReceiptsServiceStub struct {
	lock           sync.Mutex
	IssuedReceipts map[string]entities.IssueReceiptRequest
	VoidedReceipts []entities.VoidReceipt
}

func (c *ReceiptsServiceStub) IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.IssuedReceipts == nil {
		c.IssuedReceipts = make(map[string]entities.IssueReceiptRequest)
	}

	c.IssuedReceipts[request.IdempotencyKey] = request
	return entities.IssueReceiptResponse{}, nil
}

func (c *ReceiptsServiceStub) VoidReceipt(ctx context.Context, request entities.VoidReceipt) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.VoidedReceipts = append(c.VoidedReceipts, request)
	return nil
}
