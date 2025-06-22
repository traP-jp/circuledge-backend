package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
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
		CreatedAt      int32    `json:"created_at"`
		UpdatedAt      int32    `json:"updated_at"`
	}

	NoteResponse struct {
		Revision       string    `json:"revision"`
		Channel        string    `json:"channel"`
		Permission     string    `json:"permission"`
		Body           string    `json:"body"`
		ID             uuid.UUID `json:"id,omitempty" db:"id"`
		LatestRevision uuid.UUID `json:"latest_revision,omitempty" db:"latest_revision"`
		CreatedAt      int32     `json:"created_at,omitempty" db:"created_at"`
		DeletedAt      int32     `json:"deleted_at,omitempty" db:"deleted_at"`
		UpdatedAt      int32     `json:"updated_at,omitempty" db:"updated_at"`
	}

	UpdateNoteParams struct {
		Channel    uuid.UUID `json:"channel,omitempty" db:"channel"`
		Permission string    `json:"permission,omitempty" db:"permission"`
		Revision   uuid.UUID `json:"revision,omitempty" db:"revision"`
		Body       string    `json:"body,omitempty" db:"body"`
		Tags       []string  `json:"tags,omitempty" db:"tags"`
		Title      string    `json:"title,omitempty" db:"title"`
		Summary    string    `json:"summary,omitempty" db:"summary"`
	}

	NoteRevision struct {
		NoteID     uuid.UUID `json:"note_id,omitempty" db:"note_id"`
		RevisionID uuid.UUID `json:"revision_id,omitempty" db:"revision_id"`
		Channel    uuid.UUID `json:"channel,omitempty" db:"channel"`
		Permission string    `json:"permission,omitempty" db:"permission"`
		Title      string    `json:"title,omitempty" db:"title"`
		Summary    string    `json:"summary,omitempty" db:"summary"`
		Body       string    `json:"body,omitempty" db:"body"`
		UpdatedAt  time.Time `json:"updated_at,omitempty" db:"updated_at"`
	}

	GetNoteHistoryResponse struct {
		RevisionID uuid.UUID `json:"revision_id,omitempty" db:"revision_id"`
		Channel    uuid.UUID `json:"channel,omitempty" db:"channel"`
		Permission string    `json:"permission,omitempty" db:"permission"`
		UpdatedAt  int32     `json:"updated_at,omitempty" db:"updated_at"`
		Body       string    `json:"body,omitempty" db:"body"`
	}
	UserSetting struct {
		UserName       string    `json:"user_name,omitempty" db:"user_name"`
		DefaultChannel uuid.UUID `json:"default_channel,omitempty" db:"default_channel"`
	}

	GetNotesParams struct {
		Channel      string `json:"channel"`
		IncludeChild bool   `json:"includeChild"`
		Tags         []string
		Title        string `json:"title"`
		Body         string `json:"body"`
		SortKey      string `json:"sortKey"`
		Limit        int    `json:"limit"`
		Offset       int    `json:"offset"`
	}
	GetNotesResponse struct {
		ID         string   `json:"id,omitempty" db:"id"`
		Channel    string   `json:"channel,omitempty" db:"channel"`
		Permission string   `json:"permission,omitempty" db:"permission"`
		Title      string   `json:"title,omitempty" db:"title"`
		Summary    string   `json:"summary,omitempty" db:"summary"`
		Tag        []string `json:"tag,omitempty" db:"tag"`
		UpdatedAt  int32    `json:"updatedAt,omitempty" db:"updated_at"`
		CreatedAt  int32    `json:"createdAt,omitempty" db:"created_at"`
	}
)

// Todo:Delete_at
// Delete_atになにか時間が書かれたら削除されているとみなし，404を返すようにする

// GET /notes/:note-id
func (r *Repository) GetNote(ctx context.Context, noteID string) (*NoteResponse, error) {
	// Elasticsearchでnoteを検索
	res, err := r.es.Get("notes", noteID).Do(ctx) // Getメソッドを使用してドキュメントを取得
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
	// SQLのdeleted_atを更新
	query := `UPDATE notes SET deleted_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, time.Now(), noteID)
	if err != nil {
		if err.Error() == "note not found" {

			return echo.NewHTTPError(http.StatusNotFound, "note not found")
		}
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

func (r *Repository) CreateNote(ctx context.Context, channelID uuid.UUID) (uuid.UUID, uuid.UUID, string, uuid.UUID, error) {
	noteID, _ := uuid.NewV7()
	revisionID, _ := uuid.NewV7()
	permission := "limited"
	doc := map[string]interface{}{
		"id":             noteID.String(),
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
	resp, err := r.es.Index("notes").Document(doc).Id(noteID.String()).Do(ctx) // Elasticsearchにインデックス登録
	if err != nil {
		return noteID, channelID, permission, revisionID, fmt.Errorf("index user in ES: %w", err)
	}
	_ = resp

	query := `INSERT INTO notes (ID, latest_revision, created_at, deleted_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, noteID, revisionID, time.Now().Unix(), nil, time.Now().Unix())
	if err != nil {
		log.Printf("DB Error: %s", err)

		return noteID, channelID, permission, revisionID, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	query = `INSERT INTO note_revisions (note_id, revision_id, channel, permission, title, summary, body, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, noteID, revisionID, channelID, permission, "新規ノート", "新しく作成されたノート", "", time.Now().Unix())
	if err != nil {
		log.Printf("DB Error: %s", err)

		return noteID, channelID, permission, revisionID, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return noteID, channelID, permission, revisionID, nil
}

func (r *Repository) UpdateNote(ctx context.Context, noteID uuid.UUID, params UpdateNoteParams) error {
	// titleはbodyの1行目を取得する
	title := "新規ノート"
	if params.Body != "" {
		lines := strings.Split(params.Body, "\n")
		if len(lines) > 0 {
			title = strings.TrimSpace(lines[0])
		}
	}

	// summaryはbodyの最初の100文字を取得する
	summary := ""
	if params.Body != "" {
		runes := []rune(params.Body)
		if len(runes) > 100 {
			summary = string(runes[:100]) + "..." // 100文字を超える場合は省略記号を付ける
		} else {
			summary = params.Body // 100文字未満ならそのまま
		}
	}

	// tagはbodyの中にある#で始まる単語を取得することとする
	tags := []string{}
	if params.Body != "" {
		words := strings.Fields(params.Body)
		for _, word := range words {
			if strings.HasPrefix(word, "#") {
				tag := strings.TrimPrefix(word, "#")
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}
	}

	doc := map[string]interface{}{
		"channel":    params.Channel.String(),
		"permission": params.Permission,
		"revision":   params.Revision.String(),
		"body":       params.Body,
		"title":      title,
		"summary":    summary,
		"tag":        tags,
		"updatedAt":  time.Now().Unix(),
	}

	_, err := r.es.Update("notes", noteID.String()).Doc(doc).Do(ctx)
	if err != nil {

		return fmt.Errorf("update note in ES: %w", err)
	}

	// SQLからdeleted_atのみ取ってくる
	query := `SELECT deleted_at FROM notes WHERE id = ?`

	var deletedAt sql.NullTime
	err = r.db.QueryRow(query, noteID).Scan(&deletedAt)

	if err != nil {
		log.Printf("DB Error (deleted_at select): %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
	if deletedAt.Valid {

		return echo.NewHTTPError(http.StatusNotFound, "note not found")
	}

	revisionID, _ := uuid.NewV7()
	query = `UPDATE notes SET latest_revision = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, revisionID.String(), time.Now().Unix(), noteID)

	if err != nil {
		log.Printf("DB Error: %s", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	query = `INSERT INTO note_revisions (note_id, revision_id, channel, permission, title, summary, body, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, noteID, revisionID.String(), params.Channel, params.Permission, params.Title, params.Summary, params.Body, time.Now().Unix())

	if err != nil {
		log.Printf("DB Error: %s", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil
}

func (r *Repository) GetNoteHistory(_ context.Context, noteID string, limit int, offset int) ([]GetNoteHistoryResponse, error) {
	query := `SELECT revision_id, channel, permission, updated_at, body FROM note_revisions WHERE note_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	histories := []GetNoteHistoryResponse{}
	err := r.db.Select(&histories, query, noteID, limit, offset)
	if err != nil {
		log.Printf("DB Error: %s", err)

		return nil, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
	if len(histories) == 0 {

		return nil, echo.NewHTTPError(http.StatusNotFound, "not found")
	}

	return histories, nil
}

func NewTermQuery(field string, value interface{}) types.Query {
	return types.Query{
		Term: map[string]types.TermQuery{
			field: {
				Value: value,
			},
		},
	}
}

func NewMatchQuery(field string, queryText string) types.Query {
	return types.Query{
		Match: map[string]types.MatchQuery{
			field: {Query: queryText},
		},
	}
}

func NewRegexQuery(field string, pattern string) types.Query {
	return types.Query{
		Regexp: map[string]types.RegexpQuery{
			field: {Value: pattern},
		},
	}
}

func (r *Repository) GetNotes(ctx context.Context, params GetNotesParams) ([]GetNotesResponse, error) {
	var mustQueries []types.Query
	var filterQueries []types.Query
	var shouldQueries []types.Query
	if params.Channel != "" {
		shouldQueries = append(shouldQueries, NewTermQuery("channel.keyword", params.Channel))
	}
	if params.IncludeChild {
		// チャンネルの子チャンネルを取得するためのAPIを呼び出す
		// 認証ができない
		req, err := http.NewRequest("GET", "https://q.trap.jp/api/v3/channels/"+params.Channel, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for channel data: %w", err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get channel data: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get channel data: status code %d", resp.StatusCode)
		}
		var channelData struct {
			ID       string   `json:"id"`
			Children []string `json:"children"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&channelData); err != nil {
			return nil, fmt.Errorf("failed to decode channel data: %w", err)
		}
		// チャンネルの子チャンネルをフィルタに追加
		for _, childID := range channelData.Children {
			shouldQueries = append(shouldQueries, NewTermQuery("channel.keyword", childID))
		}
	}
	if params.Title != "" {
		mustQueries = append(mustQueries, NewRegexQuery("title.keyword", params.Title))
	}
	if params.Body != "" {
		mustQueries = append(mustQueries, NewRegexQuery("body.keyword", params.Body))
	}
	if len(params.Tags) > 0 {
		for _, tag := range params.Tags {
			filterQueries = append(filterQueries, NewRegexQuery("tag.keyword", tag))
		}
	}
	query := &types.Query{
		Bool: &types.BoolQuery{
			Filter: filterQueries,
			Must:   mustQueries,
			Should: shouldQueries,
		},
	}
	/*
		sort := types.Sort{}
		if params.SortKey != "" {
			switch params.SortKey {
			case "dateAsc":
				sort = types.Sort{
					&types.SortOptions{
						SortOption: map[string]types.FieldSort{
							"createdAt": {
								Order: &sortorder.Asc,
							},
						},
					},
				}
			case "dateDesc":
			case "titleAsc":
			case "titleDesc":
			default:
				return nil, fmt.Errorf("invalid sortKey value: %s", params.SortKey)
			}
		}
	*/
	res, err := r.es.Search().Index("notes").Query(query).Size(params.Limit).From(params.Offset).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search notes in ES: %w", err)
	}

	var notes []GetNotesResponse
	for _, hit := range res.Hits.Hits {
		var note GetNotesResponse
		if err := json.Unmarshal(hit.Source_, &note); err != nil {
			return nil, fmt.Errorf("unmarshal note data: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}
