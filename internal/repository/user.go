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
		ID         uuid.UUID `json:"id,omitempty" db:"id"`
		LatestRevision uuid.UUID `json:"latest_revision,omitempty" db:"latest_revision"`
		CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
		DeletedAt time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
	}

	Revision struct {
	}

	UserSetting struct {
		UserName string `json:"user_name,omitempty" db:"user_name"`
		DefaultChannel uuid.UUID `json:"default_channel,omitempty" db:"default_channel"`
	}
)

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
	channelID := uuid.New()

	query := `INSERT INTO notes (ID, latest_revision, created_at, deleted_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, noteID, revisionID, time.Now(), nil, time.Now())
	if err != nil {
		log.Printf("DB Error: %s", err)
		return noteID, channelID, permission, revisionID, echo.NewHTTPError(http.StatusInternalServerError, "internal server error1")
	}
	
	return noteID, channelID, permission, revisionID, nil
}
