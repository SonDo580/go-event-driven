package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetBookings(c echo.Context) error {
	bookings, err := h.opsBookingReadModel.AllBookings(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get all reservations: %w", err)
	}

	return c.JSON(http.StatusOK, bookings)
}

func (h Handler) GetBookingByID(c echo.Context) error {
	booking, err := h.opsBookingReadModel.ReservationReadModel(
		c.Request().Context(),
		c.Param("id"),
	)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	return c.JSON(http.StatusOK, booking)
}
