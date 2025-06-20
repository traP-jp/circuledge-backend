package repository

import (
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
	es *elasticsearch.TypedClient
}

func New(db *sqlx.DB, es *elasticsearch.TypedClient) *Repository {
	return &Repository{db: db, es: es}
}
