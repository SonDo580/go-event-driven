package constants

const (
	SvcTickets = "svc_tickets"
)

const (
	HandlerIssueReceipt         = "issue_receipt"
	HandlerAppendToTracker      = "append_to_tracker"
	HandlerPrintTicket          = "print_ticket"
	HandlerStoreTicket          = "store_ticket"
	HandlerCancelTicket         = "cancel_ticket"
	HandlerRemoveCanceledTicket = "remove_canceled_ticket"
)

const (
	TicketStatusConfirmed = "confirmed"
	TicketStatusCanceled  = "canceled"
)

const (
	SheetTicketsToPrint  = "tickets-to-print"
	SheetTicketsToRefund = "tickets-to-refund"
)

const (
	HeaderCorrelationID  = "Correlation-ID"
	HeaderIdempotencyKey = "Idempotency-Key"
)

const (
	MsgMetaCorrelationID = "correlation_id"
)
