package handler

import (
	"net/http"

	"github.com/traP-jp/circuledge-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// スキーマ定義
type (
	CreateNoteResponse struct {
		ID         string `json:"id"`
		Channel    string `json:"channel"`
		Permission string `json:"permission"`
		Revision   string `json:"revision"`
		Body       string `json:"body"`
	}

	updateNoteParams struct {
		Channel    uuid.UUID `json:"channel"`
		Permission string    `json:"permission"`
		Revision   uuid.UUID `json:"revision"`
		Body       string    `json:"body"`
	}
)

// GET /notes/:noteId
func (h *Handler) GetNote(c echo.Context) error {
	noteID := c.Param("noteId")
	if noteID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "note ID is required")
	}

	note, err := h.repo.GetNote(c.Request().Context(), noteID)
	if err != nil {
		if err.Error() == "note not found" {

			return echo.NewHTTPError(http.StatusNotFound, "note not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	res := CreateNoteResponse{
		Revision:   note.Revision,
		Channel:    note.Channel,
		Permission: note.Permission,
		Body:       note.Body,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateNote(c echo.Context) error {
	session, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session").SetInternal(err)
	}
	// normalize default channel value into a uuid.UUID
	channelUUID := uuid.Nil
	if defaultChannel := session.Values["default_channel"]; defaultChannel != nil {
		switch v := defaultChannel.(type) {
		case string:
			if parsed, err := uuid.Parse(v); err == nil {
				channelUUID = parsed
			}
		case uuid.UUID:
			channelUUID = v
		}
	}

	noteID, channelID, permission, revisionID, err := h.repo.CreateNote(c.Request().Context(), channelUUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}
	res := CreateNoteResponse{
		ID:         noteID.String(),
		Channel:    channelID.String(),
		Permission: permission,
		Revision:   revisionID.String(),
	}

	return c.JSON(http.StatusOK, res)
}
func (h *Handler) UpdateNote(c echo.Context) error {
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid note ID").SetInternal(err)
	}
	params := new(updateNoteParams)
	if err := c.Bind(params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}
	err = h.repo.UpdateNote(c.Request().Context(), noteID, repository.UpdateNoteParams{
		Channel:    params.Channel,
		Permission: params.Permission,
		Revision:   params.Revision,
		Body:       params.Body,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	return c.NoContent(http.StatusNoContent)
}
