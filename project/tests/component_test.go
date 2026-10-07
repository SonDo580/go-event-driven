package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"tickets/adapters"
	"tickets/config"
	"tickets/constants"
	ticketsDb "tickets/db"
	"tickets/entities"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/service"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	cfg := config.Get()

	db, err := sqlx.Open("postgres", cfg.PostgresUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	redisClient := message.NewRedisClient(cfg.RedisAddr)
	defer redisClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spreadsheetsAPI := &adapters.SpreadsheetsAPIStub{}
	receiptsService := &adapters.ReceiptsServiceStub{}
	filesAPI := &adapters.FileApiStub{}

	go func() {
		svc := service.New(
			db,
			redisClient,
			spreadsheetsAPI,
			receiptsService,
			filesAPI,
		)

		err := svc.Run(ctx)
		assert.NoError(t, err)
	}()

	waitForHttpServer(t)

	ticket := ticketsHttp.TicketStatusRequest{
		TicketID: watermill.NewUUID(),
		Status:   constants.TicketStatusConfirmed,
		Price: entities.Money{
			Amount:   "10.50",
			Currency: "USD",
		},
		CustomerEmail: "x@x.com",
	}

	idempotencyKey := watermill.NewUUID()

	// check idempotency
	for range 3 {
		sendTicketStatus(
			t,
			ticketsHttp.TicketsStatusRequest{
				Tickets: []ticketsHttp.TicketStatusRequest{ticket},
			},
			idempotencyKey,
		)
	}

	assertReceiptForTicketIssued(t, receiptsService, ticket)
	assertTicketPrinted(t, filesAPI, ticket)
	assertRowToSheetAdded(t, spreadsheetsAPI, ticket, constants.SheetTicketsToPrint)
	assertTicketStoredInRepository(t, db, ticket)

	ticket.Status = constants.TicketStatusCanceled
	sendTicketStatus(
		t,
		ticketsHttp.TicketsStatusRequest{
			Tickets: []ticketsHttp.TicketStatusRequest{ticket},
		},
		watermill.NewUUID(),
	)

	assertRowToSheetAdded(t, spreadsheetsAPI, ticket, constants.SheetTicketsToRefund)
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

func sendTicketStatus(
	t *testing.T,
	req ticketsHttp.TicketsStatusRequest,
	idempotencyKey string,
) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := watermill.NewShortUUID()

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set(constants.HeaderCorrelationID, correlationID)
	httpReq.Header.Set(constants.HeaderIdempotencyKey, idempotencyKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertTicketStoredInRepository(
	t *testing.T,
	db *sqlx.DB,
	ticket ticketsHttp.TicketStatusRequest,
) {
	ticketRepo := ticketsDb.NewTicketsRepository(db)

	assert.Eventually(
		t,
		func() bool {
			tickets, err := ticketRepo.FindAll(context.Background())
			if err != nil {
				return false
			}

			for _, t := range tickets {
				if t.TicketID == ticket.TicketID {
					return true
				}
			}

			return false
		},
		10*time.Second,
		100*time.Millisecond,
	)
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

			var ticketRow []string
			for _, row := range rows {
				for _, col := range row {
					if col == ticket.TicketID {
						ticketRow = row
						break
					}
				}
			}

			if !assert.NotEmpty(t, ticketRow, "ticket row not found in sheet %s", sheetName) {
				return
			}

			expectedRow := []string{
				ticket.TicketID,
				ticket.CustomerEmail,
				ticket.Price.Amount,
				ticket.Price.Currency,
			}

			assert.Equal(t, expectedRow, ticketRow)
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func assertTicketPrinted(
	t *testing.T,
	filesAPI *adapters.FileApiStub,
	ticket ticketsHttp.TicketStatusRequest,
) {
	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			fileName := ticket.TicketID + "-ticket.html"
			content, err := filesAPI.DownloadFile(context.Background(), fileName)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotEmpty(t, content) {
				return
			}

			assert.Contains(t, content, ticket.TicketID)
		},
		10*time.Second,
		100*time.Millisecond,
	)
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
			issuedReceiptsCount := len(receiptsService.IssuedReceipts)
			parentT.Log("issued receipts", issuedReceiptsCount)

			assert.Equal(t, issuedReceiptsCount, 1, "receipt for ticket %s not found", ticket.TicketID)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	receipt, ok := lo.Find(
		lo.Values(receiptsService.IssuedReceipts),
		func(r entities.IssueReceiptRequest) bool {
			return r.TicketID == ticket.TicketID
		})

	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)

	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}
