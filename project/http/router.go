package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(
	eventBus *cqrs.EventBus,
	ticketsRepo TicketsRepository,
	showsRepo ShowsRepository,
	bookingsRepo BookingsRepository,
) *echo.Echo {
	e := libHttp.NewEcho()
	// - 1 of the steps:
	//   Register a middleware that checks 'Correlation-ID' header
	//   and place the value in request context (the key is implementation details).
	// - Usage: by CorrelationPublisherDecorator

	handler := Handler{
		eventBus:     eventBus,
		ticketsRepo:  ticketsRepo,
		showsRepo:    showsRepo,
		bookingsRepo: bookingsRepo,
	}

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.POST("/tickets-status", handler.PostTicketsStatus)

	e.GET("/tickets", handler.GetTickets)

	e.POST("/shows", handler.PostShows)

	e.POST("/book-tickets", handler.PostBookings)

	return e
}
