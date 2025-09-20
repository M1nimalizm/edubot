package handlers

import (
	"net/http"
	"strconv"
	"time"

	"edubot/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AssignmentHandler представляет обработчик заданий
type AssignmentHandler struct {
	assignmentService *services.AssignmentService
}

// NewAssignmentHandler создает новый обработчик заданий
func NewAssignmentHandler(assignmentService *services.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: assignmentService,
	}
}

// CreateAssignmentRequest представляет запрос на создание задания
type CreateAssignmentRequest struct {
	Title       string      `json:"title" binding:"required"`
	Description string      `json:"description"`
	Subject     string      `json:"subject" binding:"required"`
	Deadline    time.Time   `json:"deadline" binding:"required"`
	UserIDs     []uuid.UUID `json:"user_ids"`
}

// GradeSubmissionRequest представляет запрос на оценку решения
type GradeSubmissionRequest struct {
	Grade    int    `json:"grade" binding:"required,min=1,max=5"`
	Comments string `json:"comments"`
}

// CreateAssignment создает новое задание (только для преподавателя)
func (h *AssignmentHandler) CreateAssignment(c *gin.Context) {
	var req CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID создателя из контекста
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	creatorID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	assignment, err := h.assignmentService.CreateAssignment(&services.CreateAssignmentRequest{
		Title:       req.Title,
		Description: req.Description,
		Subject:     req.Subject,
		Deadline:    req.Deadline,
		UserIDs:     req.UserIDs,
	}, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

// GetAssignments получает список заданий пользователя
func (h *AssignmentHandler) GetAssignments(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	assignments, err := h.assignmentService.GetAssignmentsForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assignments)
}

// GetAssignment получает детали задания
func (h *AssignmentHandler) GetAssignment(c *gin.Context) {
	assignmentIDStr := c.Param("id")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Получаем ID пользователя из контекста
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	assignment, err := h.assignmentService.GetAssignmentDetails(assignmentID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assignment)
}

// SubmitSolution загружает решение задания
func (h *AssignmentHandler) SubmitSolution(c *gin.Context) {
	assignmentIDStr := c.Param("id")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Получаем ID пользователя из контекста
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем файлы из формы
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided"})
		return
	}

	submission, err := h.assignmentService.SubmitSolution(assignmentID, userID, files)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, submission)
}

// GetPendingSubmissions получает непроверенные решения (только для преподавателя)
func (h *AssignmentHandler) GetPendingSubmissions(c *gin.Context) {
	submissions, err := h.assignmentService.GetPendingSubmissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

// GradeSubmission оценивает решение (только для преподавателя)
func (h *AssignmentHandler) GradeSubmission(c *gin.Context) {
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	var req GradeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.assignmentService.GradeSubmission(submissionID, req.Grade, req.Comments); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Submission graded successfully"})
}

// GetUpcomingDeadlines получает задания с приближающимися дедлайнами
func (h *AssignmentHandler) GetUpcomingDeadlines(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hours parameter"})
		return
	}

	assignments, err := h.assignmentService.GetUpcomingDeadlines(hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assignments)
}

// SendDeadlineReminders отправляет напоминания о дедлайнах (только для преподавателя)
func (h *AssignmentHandler) SendDeadlineReminders(c *gin.Context) {
	if err := h.assignmentService.SendDeadlineReminders(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deadline reminders sent successfully"})
}
