package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetBookings(c echo.Context) error {
	receiptIssueDate := c.QueryParam("receipt_issue_date")
	if receiptIssueDate != "" {
		_, err := time.Parse("2006-01-02", receiptIssueDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid receipt_issue_date format, expected RFC3339 date: ", err.Error())
		}
	}

	bookings, err := h.opsBookingReadModel.AllBookings(
		c.Request().Context(),
		receiptIssueDate,
	)
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
