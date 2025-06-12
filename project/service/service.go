package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slog"

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
	go func() {
		err := s.watermillRouter.Run(ctx)
		if err != nil {
			slog.With("error", err).Error("Failed to run watermill router")
		}
	}()

	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}
	return nil
}
