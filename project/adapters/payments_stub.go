package adapters

import (
	"context"
	"sync"
	"tickets/entities"
)

type PaymentsServiceStub struct {
	lock    sync.Mutex
	Refunds []entities.PaymentRefund
}

func (c *PaymentsServiceStub) RefundPayment(ctx context.Context, request entities.PaymentRefund) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Refunds = append(c.Refunds, request)
	return nil
}
