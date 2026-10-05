package service

import (
	"context"
	"errors"
	stdHTTP "net/http"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/message/event"
)

type Service struct {
	echoRouter      *echo.Echo
	watermillRouter *watermillMessage.Router
}

func New(
	redisClient *redis.Client,
	spreadsheetsAPI event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
) Service {
	watermillLogger := watermill.NewSlogLogger(log.FromContext(context.Background()))

	publisher := message.NewRedisPublisher(redisClient, watermillLogger)
	eventBus := event.NewBus(publisher)

	eventHandler := event.NewHandler(spreadsheetsAPI, receiptsService)
	eventProcessConfig := event.NewProcessorConfig(redisClient, watermillLogger)
	watermillRouter := message.NewWatermillRouter(
		eventHandler,
		eventProcessConfig,
		watermillLogger,
	)

	echoRouter := ticketsHttp.NewHttpRouter(eventBus)

	return Service{
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
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
