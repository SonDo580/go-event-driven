package service

import (
	"context"
	"errors"
	"fmt"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"tickets/db"
	ticketsDB "tickets/db"
	ticketsHttp "tickets/http"
	ticketsMsg "tickets/message"
	"tickets/message/event"
)

type Service struct {
	db              *sqlx.DB
	echoRouter      *echo.Echo
	watermillRouter *message.Router
}

func New(
	db *sqlx.DB,
	redisClient *redis.Client,
	spreadsheetsAPI event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
) Service {
	logger := watermill.NewSlogLogger(nil)
	publisher := ticketsMsg.NewRedisPublisher(redisClient, logger)
	eventBus := ticketsMsg.NewEventBus(publisher)

	ticketsRepo := ticketsDB.NewTicketsRepository(db)
	eventHandler := event.NewHandler(receiptsService, spreadsheetsAPI, ticketsRepo)

	echoRouter := ticketsHttp.NewHttpRouter(eventBus)
	watermillRouter := ticketsMsg.NewWatermillRouter(
		eventHandler,
		redisClient,
		logger,
	)

	return Service{
		db:              db,
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	err := db.InitializeDBSchema(s.db)
	if err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	errGrp, ctx := errgroup.WithContext(ctx)

	errGrp.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	errGrp.Go(func() error {
		// Start the HTTP server after watermill router is ready
		<-s.watermillRouter.Running()

		err := s.echoRouter.Start(":8080")
		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}
		return nil
	})

	errGrp.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(ctx)
	})

	return errGrp.Wait()
}
