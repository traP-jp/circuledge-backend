package repository

import (
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db    *sqlx.DB
	es    *elasticsearch.TypedClient
	token string
}

func New(db *sqlx.DB, es *elasticsearch.TypedClient, token string) *Repository {
	return &Repository{db: db, es: es, token: token}
}
