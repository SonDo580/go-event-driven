package db_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tickets/config"
	ticketsDb "tickets/db"
	"tickets/entities"
)

var db *sqlx.DB
var getDbOnce sync.Once

func getDb() *sqlx.DB {
	getDbOnce.Do(func() {
		var err error
		cfg := config.Get()
		db, err = sqlx.Open("postgres", cfg.PostgresUrl)
		if err != nil {
			panic(err)
		}
	})
	return db
}

func TestTicketsRepository_Add_idempotency(t *testing.T) {
	ctx := context.Background()

	db := getDb()

	err := ticketsDb.InitializeDBSchema(db)
	require.NoError(t, err)

	repo := ticketsDb.NewTicketsRepository(db)

	ticketToAdd := entities.Ticket{
		TicketID: watermill.NewUUID(),
		Price: entities.Money{
			Amount:   "1.00",
			Currency: "USD",
		},
		CustomerEmail: "x@x.com",
	}

	for range 2 {
		err := repo.Add(ctx, ticketToAdd)
		require.NoError(t, err)

		tickets, err := repo.FindAll(ctx)
		require.NoError(t, err)

		var matchedTickets []entities.Ticket
		for _, ticket := range tickets {
			if ticket.TicketID == ticketToAdd.TicketID {
				matchedTickets = append(matchedTickets, ticket)
			}
		}

		assert.Equal(t, len(matchedTickets), 1)
		assert.Equal(t, matchedTickets[0], ticketToAdd)
	}
}
