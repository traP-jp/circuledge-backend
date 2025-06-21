package handler

import (
	"fmt"
	"net/http"

	"github.com/traP-jp/circuledge-backend/internal/repository"

	vd "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// スキーマ定義
type (
	GetUsersResponse []GetUserResponse

	GetUserResponse struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
	}

	CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	CreateUserResponse struct {
		ID uuid.UUID `json:"id"`
	}

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

// GET /api/v1/users
func (h *Handler) GetUsers(c echo.Context) error {
	users, err := h.repo.GetUsers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	res := make(GetUsersResponse, len(users))
	for i, user := range users {
		res[i] = GetUserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		}
	}

	return c.JSON(http.StatusOK, res)
}

// POST /api/v1/users
func (h *Handler) CreateUser(c echo.Context) error {
	req := new(CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}

	err := vd.ValidateStruct(
		req,
		vd.Field(&req.Name, vd.Required),
		vd.Field(&req.Email, vd.Required, is.Email),
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err)).SetInternal(err)
	}

	userID, err := h.repo.CreateUser(c.Request().Context(), repository.CreateUserParams{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("invalid request body: %w", err)).SetInternal(err)
	}

	res := CreateUserResponse{
		ID: userID,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateNote(c echo.Context) error {

	noteID, channelID, permission, revisionID, err := h.repo.CreateNote(c.Request().Context())
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
