package constants

const (
	TopicTicketBookingConfirmed = "TicketBookingConfirmed"
	TopicTicketBookingCanceled  = "TicketBookingCanceled"
)

const (
	EventTypeTicketBookingConfirmed = "TicketBookingConfirmed"
	EventTypeTicketBookingCanceled  = "TicketBookingCanceled"
)

const (
	ConsumerGroupIssueReceipt    = "issue_receipt"
	ConsumerGroupAppendToTracker = "append_to_tracker"
	ConsumerGroupRefund          = "refund"
)

const (
	HandlerIssueReceipt    = "issue_receipt"
	HandlerAppendToTracker = "append_to_tracker"
	HandlerCancelTicket    = "cancel_ticket"
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
	HeaderCorrelationID = "Correlation-ID"
)

const (
	MsgMetaCorrelationID = "correlation_id"
	MsgMetaType          = "type"
)
