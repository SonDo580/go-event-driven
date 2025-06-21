package adapters

import (
	"context"
	"sync"
	"tickets/entities"
)

type ReceiptsServiceStub struct {
	lock           sync.Mutex
	IssuedReceipts []entities.IssueReceiptRequest
}

func (stub *ReceiptsServiceStub) IssueReceipt(
	ctx context.Context, request entities.IssueReceiptRequest,
) error {
	stub.lock.Lock()
	defer stub.lock.Unlock()

	stub.IssuedReceipts = append(stub.IssuedReceipts, request)
	return nil
}
