package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/pkg/config"
)

func main() {
	// Config init
	cfg := config.Load()
	db := database.Connect(cfg.DatabaseURL)
	defer db.Close()

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from Healthcare CRM backend!")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
