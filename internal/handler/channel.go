package handler

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetChannels(c echo.Context) error {
	res, err := h.repo.GetChannels()
	if err != nil {
		return echo.NewHTTPError(500, err)
	}

	return c.JSON(200, res)
}
