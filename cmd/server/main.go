package main

import (
	"github.com/gorilla/sessions"
	"github.com/traP-jp/circuledge-backend/cmd/server/server"
	"github.com/traP-jp/circuledge-backend/pkg/config"
	"github.com/traP-jp/circuledge-backend/pkg/database"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://circuledge.trap.show", "http://localhost:5173"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

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
