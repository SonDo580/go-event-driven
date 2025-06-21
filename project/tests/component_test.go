package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"tickets/adapters"
	"tickets/constants"
	"tickets/entities"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/service"

	"slices"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	redisClient := message.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer redisClient.Close()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	spreadsheetsAPI := &adapters.SpreadsheetsAPIStub{}
	receiptsService := &adapters.ReceiptsServiceStub{}

	go func() {
		err := service.New(
			redisClient,
			spreadsheetsAPI,
			receiptsService,
		).Run(ctx)
		assert.NoError(t, err)
	}()

	waitForHttpServer(t)

	confirmedTicket := ticketsHttp.TicketStatusRequest{
		TicketID:      uuid.NewString(),
		CustomerEmail: "example@example.com",
		Price: entities.Money{
			Amount:   "100",
			Currency: "USD",
		},
		Status: constants.TicketStatusConfirmed,
	}

	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{confirmedTicket},
	})
	assertReceiptForTicketIssued(t, receiptsService, confirmedTicket)
	assertRowToSheetAdded(t, spreadsheetsAPI, confirmedTicket, constants.SheetTicketsToPrint)

	canceledTicket := ticketsHttp.TicketStatusRequest{
		TicketID:      uuid.NewString(),
		CustomerEmail: "example@example.com",
		Price: entities.Money{
			Amount:   "100",
			Currency: "USD",
		},
		Status: constants.TicketStatusCanceled,
	}

	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{canceledTicket},
	})
	assertRowToSheetAdded(t, spreadsheetsAPI, canceledTicket, constants.SheetTicketsToRefund)
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}

func sendTicketsStatus(t *testing.T, req ticketsHttp.TicketsStatusRequest) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	// ticketIDs := make([]string, 0, len(req.Tickets))
	// for _, ticket := range req.Tickets {
	// 	ticketIDs = append(ticketIDs, ticket.TicketID)
	// }

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set(constants.HeaderCorrelationID, correlationID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertReceiptForTicketIssued(
	t *testing.T,
	receiptsService *adapters.ReceiptsServiceStub,
	ticket ticketsHttp.TicketStatusRequest,
) {
	t.Helper()

	parentT := t
	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			issuedReceipts := len(receiptsService.IssuedReceipts)
			parentT.Log("issued receipts", issuedReceipts)

			assert.Greater(t, issuedReceipts, 0, "no receipts issued")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var receipt entities.IssueReceiptRequest
	var ok bool
	for _, issuedReceipt := range receiptsService.IssuedReceipts {
		if issuedReceipt.TicketID != ticket.TicketID {
			continue
		}

		receipt = issuedReceipt
		ok = true
		break
	}

	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)
	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func assertRowToSheetAdded(
	t *testing.T,
	spreadsheetsAPI *adapters.SpreadsheetsAPIStub,
	ticket ticketsHttp.TicketStatusRequest,
	sheetName string,
) {
	t.Helper()

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			rows, ok := spreadsheetsAPI.Rows[sheetName]
			if !assert.True(t, ok, "sheet %s not found", sheetName) {
				return
			}

			assert.Greater(t, len(rows), 0, "no tickets found in sheet %s", sheetName)

			ok = false
			for _, row := range rows {
				if slices.Contains(row, ticket.TicketID) {
					ok = true
					break
				}
			}

			require.Truef(t, ok, "ticket %s not found in sheet %s", ticket.TicketID, sheetName)
		},
		10*time.Second,
		100*time.Millisecond,
	)

}
