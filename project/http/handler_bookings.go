package http

import (
	"errors"
	"fmt"
	"net/http"
	"tickets/db"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/labstack/echo/v4"
)

type BookingCreatedResponse struct {
	BookingID string `json:"booking_id"`
}

func (h Handler) PostBookings(c echo.Context) error {
	booking := entities.Booking{}
	err := c.Bind(&booking)
	if err != nil {
		return err
	}

	booking.BookingID = watermill.NewUUID()

	err = h.bookingsRepo.Add(c.Request().Context(), booking)
	if errors.Is(err, db.ErrOverBooking) {
		c.JSON(http.StatusBadRequest, "")
	}

	if err != nil {
		return fmt.Errorf("failed to add booking: %w", err)
	}

	return c.JSON(http.StatusCreated, BookingCreatedResponse{BookingID: booking.BookingID})
}
