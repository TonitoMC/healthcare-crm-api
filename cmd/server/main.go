package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/pkg/config"

	"github.com/tonitomc/healthcare-crm-api/internal/api/routes"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/auth"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/user"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db := database.Connect(cfg.DatabaseURL)
	defer db.Close()

	// Initialize Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Root test route
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from Healthcare CRM backend!")
	})

	// ===== Dependency Injection Setup =====

	// User dependencies
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	// Auth dependencies
	authService := auth.NewService(userService, cfg.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)

	// ===== Route Registration =====
	routes.RegisterAuthRoutes(e, authHandler)

	// ===== Server Start =====
	e.Logger.Fatal(e.Start(":8080"))
}
