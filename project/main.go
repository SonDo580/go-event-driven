package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"tickets/adapters"
	"tickets/config"
	"tickets/constants"
	"tickets/message"
	"tickets/service"
)

func main() {
	log.Init(slog.LevelInfo)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := config.Get()
	apiClients, err := clients.NewClients(
		cfg.GatewayAddr,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set(constants.HeaderCorrelationID, log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	deadNationAPI := adapters.NewDeadNationClient(apiClients)
	spreadsheetsAPI := adapters.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := adapters.NewReceiptsServiceClient(apiClients)
	filesAPI := adapters.NewFilesApiClient(apiClients)

	redisClient := message.NewRedisClient(cfg.RedisAddr)
	defer redisClient.Close()

	db, err := sqlx.Open("postgres", cfg.PostgresUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = service.New(
		db,
		redisClient,
		deadNationAPI,
		spreadsheetsAPI,
		receiptsService,
		filesAPI,
	).Run(ctx)
	if err != nil {
		panic(err)
	}
}
