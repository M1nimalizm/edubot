package handlers

import (
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"edubot/internal/models"
	"edubot/internal/services"
)

type AssignmentHandler struct {
	assignmentService *services.AssignmentService
}

func NewAssignmentHandler(assignmentService *services.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: assignmentService,
	}
}

// Request structures
type CreateAssignmentRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Subject     string    `json:"subject" binding:"required"`
	Grade       int       `json:"grade" binding:"required"`
	Level       int       `json:"level" binding:"required"`
	StudentID   uuid.UUID `json:"student_id" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
}

type UpdateAssignmentRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Subject     string    `json:"subject"`
	Grade       int       `json:"grade"`
	Level       int       `json:"level"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
}

type AddCommentRequest struct {
	Content     string `json:"content" binding:"required"`
	AuthorType  string `json:"author_type" binding:"required"`
}

type CreateContentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
}

// Assignment endpoints
func (h *AssignmentHandler) CreateAssignment(c *gin.Context) {
	var req CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Получаем ID учителя из контекста (из middleware)
	teacherID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	assignment := &models.Assignment{
		Title:       req.Title,
		Description: req.Description,
		Subject:     req.Subject,
		Grade:       req.Grade,
		Level:       req.Level,
		TeacherID:   teacherID.(uuid.UUID),
		StudentID:   req.StudentID,
		DueDate:     req.DueDate,
	}
	
	if err := h.assignmentService.CreateAssignment(assignment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, assignment)
}

func (h *AssignmentHandler) GetAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment ID"})
		return
	}
	
	assignment, err := h.assignmentService.GetAssignmentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
		return
	}
	
	c.JSON(http.StatusOK, assignment)
}

func (h *AssignmentHandler) GetStudentAssignments(c *gin.Context) {
	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	assignments, err := h.assignmentService.GetAssignmentsByStudentID(studentID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, assignments)
}

func (h *AssignmentHandler) GetTeacherAssignments(c *gin.Context) {
	teacherID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	assignments, err := h.assignmentService.GetAssignmentsByTeacherID(teacherID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, assignments)
}

func (h *AssignmentHandler) GetUpcomingDeadlines(c *gin.Context) {
	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 7
	}
	
	assignments, err := h.assignmentService.GetUpcomingDeadlines(studentID.(uuid.UUID), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, assignments)
}

func (h *AssignmentHandler) UpdateAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment ID"})
		return
	}
	
	var req UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	assignment, err := h.assignmentService.GetAssignmentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
		return
	}
	
	// Обновляем поля
	if req.Title != "" {
		assignment.Title = req.Title
	}
	if req.Description != "" {
		assignment.Description = req.Description
	}
	if req.Subject != "" {
		assignment.Subject = req.Subject
	}
	if req.Grade != 0 {
		assignment.Grade = req.Grade
	}
	if req.Level != 0 {
		assignment.Level = req.Level
	}
	if !req.DueDate.IsZero() {
		assignment.DueDate = req.DueDate
	}
	if req.Status != "" {
		assignment.Status = req.Status
	}
	
	if err := h.assignmentService.UpdateAssignment(assignment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, assignment)
}

func (h *AssignmentHandler) MarkAssignmentCompleted(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment ID"})
		return
	}
	
	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	if err := h.assignmentService.MarkAssignmentCompleted(id, studentID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "assignment marked as completed"})
}

func (h *AssignmentHandler) DeleteAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment ID"})
		return
	}
	
	teacherID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	if err := h.assignmentService.DeleteAssignment(id, teacherID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "assignment deleted"})
}

// Comment endpoints
func (h *AssignmentHandler) AddComment(c *gin.Context) {
	assignmentIDStr := c.Param("id")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment ID"})
		return
	}
	
	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	comment := &models.Comment{
		Content:       req.Content,
		AuthorType:    req.AuthorType,
		AssignmentID: assignmentID,
		AuthorID:      userID.(uuid.UUID),
	}
	
	if err := h.assignmentService.AddComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, comment)
}

func (h *AssignmentHandler) GetComments(c *gin.Context) {
	assignmentIDStr := c.Param("id")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment ID"})
		return
	}
	
	comments, err := h.assignmentService.GetCommentsByAssignmentID(assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, comments)
}

// Content endpoints
func (h *AssignmentHandler) CreateContent(c *gin.Context) {
	var req CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	teacherID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	content := &models.Content{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Category:    req.Subject, // Используем Subject как Category
		CreatedBy:   teacherID.(uuid.UUID),
		IsPublic:    true,
	}
	
	if err := h.assignmentService.CreateContent(content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, content)
}

func (h *AssignmentHandler) GetContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content ID"})
		return
	}
	
	content, err := h.assignmentService.GetContentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
		return
	}
	
	c.JSON(http.StatusOK, content)
}

func (h *AssignmentHandler) GetContentBySubject(c *gin.Context) {
	subject := c.Param("subject")
	gradeStr := c.Param("grade")
	grade, err := strconv.Atoi(gradeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid grade"})
		return
	}
	
	content, err := h.assignmentService.GetContentBySubject(subject, grade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, content)
}

func (h *AssignmentHandler) GetTeacherContent(c *gin.Context) {
	teacherID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	content, err := h.assignmentService.GetContentByTeacherID(teacherID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, content)
}

func (h *AssignmentHandler) UpdateContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content ID"})
		return
	}
	
	var req CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	content, err := h.assignmentService.GetContentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
		return
	}
	
	// Обновляем поля
	if req.Title != "" {
		content.Title = req.Title
	}
	if req.Description != "" {
		content.Description = req.Description
	}
	if req.Type != "" {
		content.Type = req.Type
	}
	if req.Subject != "" {
		content.Category = req.Subject // Используем Subject как Category
	}
	
	if err := h.assignmentService.UpdateContent(content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, content)
}

func (h *AssignmentHandler) DeleteContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content ID"})
		return
	}
	
	teacherID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	if err := h.assignmentService.DeleteContent(id, teacherID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "content deleted"})
}

// Progress endpoints
func (h *AssignmentHandler) GetStudentProgress(c *gin.Context) {
	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	
	progress, err := h.assignmentService.GetStudentProgress(studentID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, progress)
}