package handler

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type UpdateSettingsParams struct {
	DefaultChannel string `json:"defaultChannel"`
}

func (h *Handler) UpdateSettings(c echo.Context) error {
	session, err := session.Get("session", c)
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	if err != nil {
		return echo.NewHTTPError(500, "failed to get session").SetInternal(err)
	}

	settings := new(UpdateSettingsParams)
	if err := c.Bind(settings); err != nil {
		return echo.NewHTTPError(400, "invalid request body").SetInternal(err)
	}
	session.Values["default_channel"] = settings.DefaultChannel

	if err := session.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(500, "failed to save session").SetInternal(err)
	}

	return c.NoContent(204)
}
