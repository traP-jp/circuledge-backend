package server

import (
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/traP-jp/circuledge-backend/internal/handler"
	"github.com/traP-jp/circuledge-backend/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type Server struct {
	handler *handler.Handler
}

func Inject(db *sqlx.DB) *Server {
	espwd, _ := os.LookupEnv("ELASTIC_PASSWORD")
	es, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch:9200"},
		Username:  "elastic",
		Password:  espwd,
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	token, _ := os.LookupEnv("BOT_ACCESS_TOKEN")

	repo := repository.New(db, es, token)
	h := handler.New(repo)

	return &Server{
		handler: h,
	}
}

func (d *Server) SetupRoutes(g *echo.Group) {
	// TODO: handler.SetupRoutesを呼び出す or 直接書く？
	d.handler.SetupRoutes(g)
}
