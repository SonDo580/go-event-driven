package http

import (
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/labstack/echo/v4"
)

type ShowCreatedResponse struct {
	ShowID string `json:"show_id"`
}

func (h Handler) PostShows(c echo.Context) error {
	show := entities.Show{}
	err := c.Bind(&show)
	if err != nil {
		return err
	}

	show.ShowID = watermill.NewUUID()

	err = h.showsRepo.Add(c.Request().Context(), show)
	if err != nil {
		return fmt.Errorf("failed to add show: %w", err)
	}

	return c.JSON(http.StatusCreated, ShowCreatedResponse{ShowID: show.ShowID})
}
