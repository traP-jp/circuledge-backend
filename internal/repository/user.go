package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
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
    ID             string    `json:"id"`
    LatestRevision string    `json:"latest_revision"`
    Channel        string    `json:"channel"`
    Permission     string    `json:"permission"`
    Title          string    `json:"title"`
    Summary        string    `json:"summary,omitempty"`
    Body           string    `json:"body"`
    Tag            []string  `json:"tag"`
    CreatedAt      string    `json:"created_at"`
    UpdatedAt      string    `json:"updated_at"`
	}

	NoteResponse struct {
		Revision string `json:"revision"`
		Channel string `json:"channel"`
		Permission string `json:"permission"`
		Body string `json:"body"`
	}
)

func intPtr(i int) *int { return &i }

// GET /notes/:note-id
// GET /notes/:note-id
func (r *Repository) GetNote(ctx context.Context, noteID string) (*NoteResponse, error) {
    // Elasticsearchでnoteを検索
    searchReq := r.es.Search().
	Index("notes").
	Query(&types.Query{
		Term: map[string]types.TermQuery{
                "id": {Value: noteID},
            },
	}).
	Size(1)

    res, err := searchReq.Do(ctx)
    if err != nil {
        return nil, fmt.Errorf("search note in ES: %w", err)
    }
    if len(res.Hits.Hits) == 0 {
        return nil, fmt.Errorf("note not found")
    }
    var note Note
    if err := json.Unmarshal(res.Hits.Hits[0].Source_, &note); err != nil {
        return nil, fmt.Errorf("unmarshal note data: %w", err)
    }
    return &NoteResponse{
        Revision:   note.LatestRevision,
        Channel:    note.Channel,
        Permission: note.Permission,
        Body:       note.Body,
    }, nil
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