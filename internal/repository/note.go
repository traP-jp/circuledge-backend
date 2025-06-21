package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type (
	// users table
	User struct {
		ID    uuid.UUID
		Name  string
		Email string
	}

	CreateUserParams struct {
		Name  string
		Email string
	}

	UpdateUserParams struct {
		ID    uuid.UUID
		Name  string
		Email string
	}

	Note struct {
		ID             string   `json:"id"`
		LatestRevision string   `json:"latestRevision"`
		Channel        string   `json:"channel"`
		Permission     string   `json:"permission"`
		Title          string   `json:"title"`
		Summary        string   `json:"summary"`
		Body           string   `json:"body"`
		Tag            []string `json:"tag"`
		CreatedAt      int      `json:"created_at"`
		UpdatedAt      int      `json:"updated_at"`
	}

	NoteResponse struct {
		Revision       string    `json:"revision"`
		Channel        string    `json:"channel"`
		Permission     string    `json:"permission"`
		Body           string    `json:"body"`
		ID             uuid.UUID `json:"id,omitempty" db:"id"`
		LatestRevision uuid.UUID `json:"latest_revision,omitempty" db:"latest_revision"`
		CreatedAt      time.Time `json:"created_at,omitempty" db:"created_at"`
		DeletedAt      time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
		UpdatedAt      time.Time `json:"updated_at,omitempty" db:"updated_at"`
	}

	UpdateNoteParams struct {
		Channel    uuid.UUID `json:"channel,omitempty" db:"channel"`
		Permission string    `json:"permission,omitempty" db:"permission"`
		Revision   uuid.UUID `json:"revision,omitempty" db:"revision"`
		Body       string    `json:"body,omitempty" db:"body"`
	}

	Revision struct {
	}

	UserSetting struct {
		UserName       string    `json:"user_name,omitempty" db:"user_name"`
		DefaultChannel uuid.UUID `json:"default_channel,omitempty" db:"default_channel"`
	}
)

// GET /notes/:note-id
func (r *Repository) GetNote(ctx context.Context, noteID string) (*NoteResponse, error) {
	// Elasticsearchでnoteを検索
	res, err := r.es.Get("notes", noteID).Do(ctx) // Getメソッドを使用してドキュメントを取得
	fmt.Println("res:", res)
	if err != nil {
		return nil, fmt.Errorf("search note in ES: %w", err)
	}
	if !res.Found {
		return nil, fmt.Errorf("note not found")
	}
	var note Note

	if err := json.Unmarshal(res.Source_, &note); err != nil {
		return nil, fmt.Errorf("unmarshal note data: %w", err)
	}

	return &NoteResponse{
		Revision:   note.LatestRevision,
		Channel:    note.Channel,
		Permission: note.Permission,
		Body:       note.Body,
	}, nil
}

// DELETE /notes/:note-id
func (r *Repository) DeleteNote(ctx context.Context, noteID string) error {
	// Elasticsearchからノートを削除
	_, err := r.es.Delete("notes", noteID).Do(ctx)
	if err != nil {

		return fmt.Errorf("delete note in ES: %w", err)
	}
	// DBからも削除
	query := `DELETE FROM notes WHERE id = ?`
	_, err = r.db.Exec(query, noteID)
	if err != nil {
		log.Printf("DB Error: %s", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
	return nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	searchReq := r.es.Search().Index("users").Size(1000)
	res, err := searchReq.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search users in ES: %w", err)
	}

	for _, hit := range res.Hits.Hits {
		var user User
		if err := json.Unmarshal(hit.Source_, &user); err != nil {
			return nil, fmt.Errorf("unmarshal user data: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (uuid.UUID, error) {
	userID := uuid.New()

	doc := map[string]interface{}{
		"id":    userID.String(),
		"name":  params.Name,
		"email": params.Email,
	}

	resp, err := r.es.Index("users").Document(doc).Id(userID.String()).Do(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("index user in ES: %w", err)
	}
	_ = resp

	return userID, nil
}

func (r *Repository) CreateNote(ctx context.Context) (uuid.UUID, uuid.UUID, string, uuid.UUID, error) {
	noteID, _ := uuid.NewV7()
	revisionID, _ := uuid.NewV7()
	permission := "limited"
	channelID := uuid.New() //todo

	doc := map[string]interface{}{
		"latestRevision": revisionID.String(),
		"channel":        channelID.String(),
		"permission":     permission,
		"title":          "新規ノート",
		"summary":        "新しく作成されたノート",
		"body":           "",
		"tag":            []string{},
		"createdAt":      time.Now().Unix(),
		"updatedAt":      time.Now().Unix(),
	}
	log.Printf("doc: %v", doc)
	resp, err := r.es.Index("notes").Document(doc).Id(noteID.String()).Do(ctx)
	if err != nil {
		return noteID, channelID, permission, revisionID, fmt.Errorf("index user in ES: %w", err)
	}
	_ = resp

	query := `INSERT INTO notes (ID, latest_revision, created_at, deleted_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, noteID, revisionID, time.Now(), nil, time.Now())
	if err != nil {
		log.Printf("DB Error: %s", err)

		return noteID, channelID, permission, revisionID, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return noteID, channelID, permission, revisionID, nil
}
func (r *Repository) UpdateNote(ctx context.Context, noteID uuid.UUID, params UpdateNoteParams) error {
	doc := map[string]interface{}{
		"channel":    params.Channel.String(),
		"permission": params.Permission,
		"revision":   params.Revision.String(),
		"body":       params.Body,
		"updatedAt":  time.Now().Unix(),
	}

	_, err := r.es.Update("notes", noteID.String()).Doc(doc).Do(ctx)
	if err != nil {
		return fmt.Errorf("update note in ES: %w", err)
	}

	query := `UPDATE notes SET body = ?, permission = ?, revision = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, params.Body, params.Permission, params.Revision.String(), time.Now(), noteID)
	if err != nil {
		log.Printf("DB Error: %s", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil
}
