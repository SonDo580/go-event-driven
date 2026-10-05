package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(eventBus *cqrs.EventBus) *echo.Echo {
	e := libHttp.NewEcho()
	// - 1 of the steps:
	//   Register a middleware that checks 'Correlation-ID' header
	//   and place the value in request context (the key is implementation details).
	// - Usage: by CorrelationPublisherDecorator

	handler := Handler{eventBus: eventBus}

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.POST("/tickets-status", handler.PostTicketsStatus)

	return e
}
