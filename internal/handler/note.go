package handler

import (
	"net/http"
	"strconv"
	"github.com/traP-jp/circuledge-backend/internal/repository"

	"github.com/google/uuid"
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
		Tags	   []string  `json:"tags"`
		Title      string    `json:"title"`
		Summary    string    `json:"summary"`
	}

	GetNoteHistoryResponse struct {
		Total int64  `json:"total"`
		Notes []repository.GetNoteHistoryResponse `json:"notes"`
	}

	GetNotesResponse struct {
		Total int64                     `json:"total"`
		Notes []repository.GetNotesResponse `json:"notes"`
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
		Tags:       params.Tags,
		Title:      params.Title,
		Summary:    params.Summary,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetNoteHistory(c echo.Context) error {
	noteID := c.Param("noteId")
	if noteID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "note ID is required")
	}
	limitStr := c.QueryParam("limit")
	offsetStr:= c.QueryParam("offset")
	if limitStr == "" {
		limitStr = "100" // Default limit
	}
	if offsetStr == "" {
		offsetStr = "0" // Default offset
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {

		return echo.NewHTTPError(http.StatusBadRequest, "invalid limit value")
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {

		return echo.NewHTTPError(http.StatusBadRequest, "invalid offset value")
	}
	histories, err := h.repo.GetNoteHistory(c.Request().Context(), noteID, limit, offset)
	if err != nil {
		if err.Error() == "note not found" {

			return echo.NewHTTPError(http.StatusNotFound, "note not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	return c.JSON(http.StatusOK, GetNoteHistoryResponse{
		Total: int64(len(histories)),
		Notes: histories,
	})
}

func (h *Handler) GetNotes(c echo.Context) error {
	channel := c.QueryParam("channel")
	if channel == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "channel ID is required")
	}
		
	inCludeChildStr := c.QueryParam("includeChild")
	if inCludeChildStr != "" && inCludeChildStr != "true" && inCludeChildStr != "false" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid includeChild value")
	}
	if inCludeChildStr == "" {
		inCludeChildStr = "false" // Default value
	}
	includeChild, err := strconv.ParseBool(inCludeChildStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid includeChild value").SetInternal(err)
	}
	tags := c.QueryParams()["tag"]
	title := c.QueryParam("title")
	body := c.QueryParam("body")
	sortkey := c.QueryParam("sortKey")
	if sortkey != "" && sortkey != "dateAsc" && sortkey != "dateDesc" && sortkey != "titleAsc" && sortkey != "titleDesc" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid sortKey value")
	}
	if sortkey == "" {
		sortkey = "dateDesc" // Default sort key
	}
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")
	if limitStr == "" {
		limitStr = "100" // Default limit
	}
	if offsetStr == "" {
		offsetStr = "0" // Default offset
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid limit value")
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid offset value")
	}
	params := repository.GetNotesParams{
		Channel:       channel,
		IncludeChild:  includeChild,
		Tags:          tags,
		Title:         title,
		Body:          body,
		SortKey:       sortkey,
		Limit:         limit,
		Offset:        offset,
	}
	notes, err := h.repo.GetNotes(c.Request().Context(), params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	return c.JSON(http.StatusOK, GetNotesResponse{
		Total: int64(len(notes)),
		Notes: notes,
	})
}