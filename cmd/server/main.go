package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tonitomc/healthcare-crm-api/internal/adapters"
	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/pkg/config"

	middlewarePkg "github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/api/routes"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/auth"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/consultation"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/medicalrecord"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/questionnaire"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/rbac"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/reminder"
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

	s3Cfg := adapters.S3Config{
		Bucket:         cfg.S3Bucket,
		Region:         cfg.S3Region,
		Endpoint:       cfg.S3Endpoint,
		AccessKey:      cfg.S3AccessKey,
		SecretKey:      cfg.S3SecretKey,
		ForcePathStyle: cfg.S3ForcePathStyle,
	}

	s3Adapter, err := adapters.NewS3Adapter(s3Cfg)
	if err != nil {
		log.Fatalf("Failed to initialize S3/MinIO adapter: %v", err)
	}

	// Initialize Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Use(middlewarePkg.JWTMiddleware(cfg.JWTSecret))

	// Root test route
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from Healthcare CRM backend!")
	})

	// ===== Dependency Injection Setup =====

	// Role dependencies
	roleRepo := role.NewRepository(db)
	roleService := role.NewService(roleRepo)
	roleHandler := role.NewHandler(roleService)

	// User dependencies
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, roleService)
	userHandler := user.NewHandler(userService)

	userPermAdapter := adapters.NewUserPermissionAdapter(userService)
	middlewarePkg.InjectPermissionProvider(userPermAdapter)

	rbacService := rbac.NewService(userService, roleService)

	// Auth Config
	authCfg := auth.Config{
		JWTSecret: cfg.JWTSecret,
		AccessTTL: cfg.JWTTTL,
		Issuer:    cfg.JWTIssuer,
	}

	// Auth dependencies
	authService := auth.NewService(userService, rbacService, authCfg)
	authHandler := auth.NewHandler(authService)

	ensureSuperuser(cfg, userService, authService, e.Logger)
	ensureSecretary(cfg, userService, authService, e.Logger)

	// Schedule dependencies
	scheduleRepo := schedule.NewRepository(db)
	scheduleService := schedule.NewService(scheduleRepo)
	scheduleHandler := schedule.NewHandler(scheduleService)

	// Patient dependencies, handler declared further down
	// as it works as an orchestration layer for response enrichment
	patientRepo := patient.NewRepository(db)
	patientService := patient.NewService(patientRepo)

	patientProvider := &adapters.PatientAdapter{Service: patientService}
	// MedicalRecord dependencies
	recordRepo := medicalrecord.NewRepository(db)
	recordService := medicalrecord.NewService(recordRepo)
	recordHandler := medicalrecord.NewHandler(recordService)

	// Questionnaire dependencies
	questionnaireRepo := questionnaire.NewRepository(db)
	questionnaireService := questionnaire.NewService(questionnaireRepo)
	questionnaireHandler := questionnaire.NewHandler(questionnaireService)

	questionnaireValidator := &adapters.QuestionnaireAdapter{Service: questionnaireService}

	// Consultation dependencies
	consultationRepo := consultation.NewRepository(db)
	consultationService := consultation.NewService(consultationRepo, questionnaireValidator)
	consultationHandler := consultation.NewHandler(consultationService)

	// Exam dependencies
	examRepo := exam.NewRepository(db)
	examService := exam.NewService(examRepo, patientProvider, s3Adapter)
	examHandler := exam.NewHandler(examService)

	// Adapters para appointments
	patientAdapter := adapters.NewPatientAdapter(patientService)
	scheduleAdapter := adapters.NewScheduleAdapter(scheduleService)

	// Appointment dependencies
	appointmentRepo := appointment.NewRepository(db)
	appointmentService := appointment.NewService(appointmentRepo, patientAdapter, scheduleAdapter)
	appointmentHandler := appointment.NewHandler(appointmentService)
	patientHandler := patient.NewHandler(patientService, examService, consultationService, recordService)

	// Reminder dependencies
	reminderRepo := reminder.NewRepository(db)
	reminderService := reminder.NewService(reminderRepo)
	reminderHandler := reminder.NewHandler(reminderService)

	// Health check route
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"status": "ok",
		})
	})

	// ===== Route Registration =====
	routes.RegisterRoutes(e, recordHandler, reminderHandler, authHandler, scheduleHandler, userHandler, roleHandler, patientHandler, consultationHandler, examHandler, appointmentHandler, questionnaireHandler)

	// ===== Server Start =====
	e.Logger.Fatal(e.Start(":8080"))
}

// Superuser bootstrap, I have NO clue where to drop this so it's here for now
func ensureSuperuser(cfg *config.Config, userService user.Service, authService auth.Service, logger echo.Logger) {
	if cfg.SuperuserEmail == "" || cfg.SuperuserPassword == "" {
		logger.Info("Skipping superuser creation — SUPERUSER_* variables not set.")
		return
	}

	user, err := userService.GetByUsernameOrEmail(cfg.SuperuserEmail)
	if err == nil && user != nil {
		logger.Infof("Superuser '%s' already exists.", cfg.SuperuserEmail)
		return
	}

	if err := authService.Register(cfg.SuperuserName, cfg.SuperuserEmail, cfg.SuperuserPassword); err != nil {
		logger.Errorf("Failed to register superuser: %v", err)
		return
	}

	user, err = userService.GetByUsernameOrEmail(cfg.SuperuserName)
	if err != nil {
		logger.Errorf("Superuser wasn't registered correctly: %v", err)
	}

	if err := userService.AddRole(user.ID, 3); err != nil {
		logger.Errorf("Failed to add Admin role to superuser: %v", err)
	}

	logger.Infof("Superuser '%s' created successfully.", cfg.SuperuserEmail)
}

// Secretary bootstrap — ensures a default secretary account exists
func ensureSecretary(cfg *config.Config, userService user.Service, authService auth.Service, logger echo.Logger) {
	if cfg.SecretaryEmail == "" || cfg.SecretaryPassword == "" {
		logger.Info("Skipping secretary creation — SECRETARY_* variables not set.")
		return
	}

	user, err := userService.GetByUsernameOrEmail(cfg.SecretaryEmail)
	if err == nil && user != nil {
		logger.Infof("Secretary '%s' already exists.", cfg.SecretaryEmail)
		return
	}

	if err := authService.Register(cfg.SecretaryName, cfg.SecretaryEmail, cfg.SecretaryPassword); err != nil {
		logger.Errorf("Failed to register secretary: %v", err)
		return
	}

	user, err = userService.GetByUsernameOrEmail(cfg.SecretaryName)
	if err != nil {
		logger.Errorf("Secretary wasn't registered correctly: %v", err)
	}

	// ⚠️ Adjust role ID here if needed — assuming 2 is the secretary role
	if err := userService.AddRole(user.ID, 2); err != nil {
		logger.Errorf("Failed to add Secretary role to account: %v", err)
	}

	logger.Infof("Secretary '%s' created successfully.", cfg.SecretaryEmail)
}
