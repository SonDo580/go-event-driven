package service

import (
	"context"
	"errors"
	stdHTTP "net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	ticketsHttp "tickets/http"
	"tickets/message"
)

type Service struct {
	echoRouter *echo.Echo
}

func New(
	redisClient *redis.Client,
	spreadsheetsAPI message.SpreadsheetsAPI,
	receiptsService message.ReceiptsService,
) Service {
	logger := watermill.NewSlogLogger(nil)
	publisher := message.NewRedisPublisher(redisClient, logger)
	echoRouter := ticketsHttp.NewHttpRouter(publisher)

	handlers := message.NewHandlers(receiptsService, spreadsheetsAPI, redisClient, logger)
	for _, handler := range handlers {
		go handler()
	}

	return Service{
		echoRouter: echoRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	err := s.echoRouter.Start(":8080")
	if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
		return err
	}
	return nil
}
