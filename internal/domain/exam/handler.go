package exam

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam/models"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(e *echo.Group) {
	exams := e.Group("/exams")
	exams.GET("/:id", h.GetByID, middleware.RequirePermission("ver-examenes"))
	exams.GET("/:id/download", h.DownloadPDF, middleware.RequirePermission("ver-examenes"))
	exams.GET("/patient/:patientId", h.GetByPatient, middleware.RequirePermission("ver-examenes"))
	exams.POST("", h.Create, middleware.RequirePermission("crear-examenes"))
	exams.POST("/:id/upload", h.UploadPDF, middleware.RequirePermission("editar-examenes"))
	exams.DELETE("/:id", h.Delete, middleware.RequirePermission("eliminar-examenes"))
	exams.GET("/pending", h.GetPending, middleware.RequirePermission("ver-examenes"))
	exams.GET("/completed", h.GetCompleted, middleware.RequirePermission("ver-examenes"))
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}

	exam, err := h.service.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, exam)
}

func (h *Handler) GetByPatient(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid patient ID"})
	}

	exams, err := h.service.GetByPatient(patientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, exams)
}

func (h *Handler) Create(c echo.Context) error {
	var req models.ExamCreateDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	id, err := h.service.Create(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{"id": id})
}

func (h *Handler) UploadPDF(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}

	// Get file from multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "File is required"})
	}

	// Validate file type
	if file.Header.Get("Content-Type") != "application/pdf" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Only PDF files are allowed"})
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "/app/uploads/exams"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create upload directory"})
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", id, file.Filename)
	filePath := filepath.Join(uploadsDir, filename)

	// Save file to disk
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to open uploaded file"})
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create file"})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to save file"})
	}

	// Update database with file info
	uploadDTO := &models.ExamUploadDTO{
		S3Key:    filename, // Store filename instead of full path
		FileSize: file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	if err := h.service.Update(id, uploadDTO); err != nil {
		// Clean up file if database update fails
		os.Remove(filePath)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message":  "PDF uploaded successfully",
		"filename": filename,
	})
}

func (h *Handler) DownloadPDF(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}

	// Get exam details from database
	exam, err := h.service.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	// Check if file exists
	if exam.S3Key == nil || *exam.S3Key == "" {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "No file found for this exam"})
	}

	// Construct file path
	filePath := filepath.Join("/app/uploads/exams", *exam.S3Key)

	// Check if file exists on disk
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "File not found on disk"})
	}

	// Serve the file
	return c.File(filePath)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}

	if err := h.service.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Exam deleted successfully"})
}

func (h *Handler) GetPending(c echo.Context) error {
	exams, err := h.service.GetPending()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, exams)
}

func (h *Handler) GetCompleted(c echo.Context) error {
	exams, err := h.service.GetCompleted()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, exams)
}
