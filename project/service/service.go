package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	ticketsHttp "tickets/http"
	ticketsMsg "tickets/message"
)

type Service struct {
	echoRouter      *echo.Echo
	watermillRouter *message.Router
}

func New(
	redisClient *redis.Client,
	spreadsheetsAPI ticketsMsg.SpreadsheetsAPI,
	receiptsService ticketsMsg.ReceiptsService,
) Service {
	logger := watermill.NewSlogLogger(nil)
	publisher := ticketsMsg.NewRedisPublisher(redisClient, logger)
	echoRouter := ticketsHttp.NewHttpRouter(publisher)
	watermillRouter := ticketsMsg.NewWatermillRouter(receiptsService, spreadsheetsAPI, redisClient, logger)

	return Service{
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
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
