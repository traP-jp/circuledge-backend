package main

import (
	"github.com/traP-jp/h25w_16_practice/cmd/server/server"
	"github.com/traP-jp/h25w_16_practice/pkg/config"
	"github.com/traP-jp/h25w_16_practice/pkg/database"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// connect to and migrate database
	db, err := database.Setup(config.MySQL())
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	s := server.Inject(db)

	v1API := e.Group("/api/v1")
	s.SetupRoutes(v1API)

	e.Logger.Fatal(e.Start(config.AppAddr()))
}
