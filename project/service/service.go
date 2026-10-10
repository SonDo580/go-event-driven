package service

import (
	"context"
	"errors"
	stdHTTP "net/http"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"tickets/db"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/message/event"
	"tickets/message/outbox"
)

type Service struct {
	db              *sqlx.DB
	echoRouter      *echo.Echo
	watermillRouter *watermillMessage.Router
}

func New(
	dbConn *sqlx.DB,
	redisClient *redis.Client,
	spreadsheetsAPI event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
	filesAPI event.FilesAPI,
) Service {
	ticketsRepo := db.NewTicketsRepository(dbConn)
	showsRepo := db.NewShowsRepository(dbConn)
	bookingsRepo := db.NewBookingsRepository(dbConn)

	watermillLogger := watermill.NewSlogLogger(log.FromContext(context.Background()))

	redisPub := message.NewRedisPublisher(redisClient, watermillLogger)
	eventBus := event.NewBus(redisPub)

	postgresSub := outbox.NewPostgresSubscriber(dbConn, watermillLogger)

	eventHandler := event.NewHandler(
		spreadsheetsAPI,
		receiptsService,
		filesAPI,
		ticketsRepo,
		eventBus,
	)
	eventProcessorConfig := event.NewProcessorConfig(redisClient, watermillLogger)
	watermillRouter := message.NewWatermillRouter(
		postgresSub,
		redisPub,
		eventProcessorConfig,
		eventHandler,
		watermillLogger,
	)

	echoRouter := ticketsHttp.NewHttpRouter(
		eventBus,
		ticketsRepo,
		showsRepo,
		bookingsRepo,
	)

	return Service{
		db:              dbConn,
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	if err := db.InitializeDBSchema(s.db); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	g.Go(func() error {
		<-s.watermillRouter.Running()

		err := s.echoRouter.Start(":8080")
		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		return s.echoRouter.Shutdown(shutdownCtx)
	})

	return g.Wait()
}
