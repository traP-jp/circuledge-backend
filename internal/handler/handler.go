package handler

import (
	"github.com/traP-jp/circuledge-backend/internal/repository"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) SetupRoutes(api *echo.Group) {
	// ping API
	pingAPI := api.Group("/ping")
	{
		pingAPI.GET("", h.Ping)
	}

	// user API
	userAPI := api.Group("/users")
	{
		userAPI.GET("", h.GetUsers)
		userAPI.POST("", h.CreateUser)
	}

	noteAPI := api.Group("/notes")
	{
		noteAPI.POST("", h.CreateNote)
		noteAPI.PUT("/:id", h.UpdateNote)
	}
}
