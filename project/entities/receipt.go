package entities

import "time"

type IssueReceiptRequest struct {
	TicketID       string `json:"ticket_id"`
	Price          Money  `json:"price"`
	IdempotencyKey string `json:"idempotency_key"`
}

type IssueReceiptResponse struct {
	ReceiptNumber string
	IssuedAt      time.Time
}

type VoidReceipt struct {
	TicketID       string
	Reason         string
	IdempotencyKey string
}
