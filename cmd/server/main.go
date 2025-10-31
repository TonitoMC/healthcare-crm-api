package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/pkg/config"

	"github.com/tonitomc/healthcare-crm-api/internal/api/routes"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/auth"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/rbac"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/role"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule"
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

	// Role dependencies
	roleRepo := role.NewRepository(db)
	roleService := role.NewService(roleRepo)

	// User dependencies
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, roleService)

	rbacService := rbac.NewService(userService, roleService)

	// Auth Config (from pkg/config)
	authCfg := auth.Config{
		JWTSecret: cfg.JWTSecret,
		AccessTTL: 24 * time.Hour, // or cfg.JWT_TTL if you added it in config.Load()
		Issuer:    "healthcare-crm",
	}

	// Auth dependencies
	authService := auth.NewService(userService, rbacService, authCfg)
	authHandler := auth.NewHandler(authService)

	// Schedule dependencies
	scheduleRepo := schedule.NewRepository(db)
	scheduleService := schedule.NewService(scheduleRepo)
	scheduleHandler := schedule.NewHandler(scheduleService)

	// ===== Route Registration =====
	routes.RegisterRoutes(e, authHandler, scheduleHandler)

	// ===== Server Start =====
	e.Logger.Fatal(e.Start(":8080"))
}
