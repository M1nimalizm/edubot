package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/services"
)

type StudentHandler struct {
	assignmentService   services.AssignmentService
	submissionService   services.SubmissionService
	gradingService      services.GradingService
	chatService         services.ChatService
	notificationService services.NotificationService
}

func NewStudentHandler(
	assignmentService services.AssignmentService,
	submissionService services.SubmissionService,
	gradingService services.GradingService,
	chatService services.ChatService,
	notificationService services.NotificationService,
) *StudentHandler {
	return &StudentHandler{
		assignmentService:   assignmentService,
		submissionService:   submissionService,
		gradingService:      gradingService,
		chatService:         chatService,
		notificationService: notificationService,
	}
}

// GET /api/student/assignments - Получить список заданий студента
func (h *StudentHandler) GetAssignments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем параметры запроса
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Получаем задания студента
	targets, err := h.assignmentService.GetAssignmentTargetsByStudent(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignments"})
		return
	}

	// Фильтруем по статусу, если указан
	var filteredTargets []*models.AssignmentTarget
	if status != "" {
		for _, target := range targets {
			if string(target.Status) == status {
				filteredTargets = append(filteredTargets, target)
			}
		}
	} else {
		filteredTargets = targets
	}

	// Применяем пагинацию
	start := offset
	end := offset + limit
	if start >= len(filteredTargets) {
		filteredTargets = []*models.AssignmentTarget{}
	} else if end > len(filteredTargets) {
		end = len(filteredTargets)
	}

	result := filteredTargets[start:end]

	c.JSON(http.StatusOK, gin.H{
		"assignments": result,
		"total":       len(filteredTargets),
		"limit":       limit,
		"offset":      offset,
	})
}

// GET /api/student/assignments/:id - Получить детали задания
func (h *StudentHandler) GetAssignment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Получаем AssignmentTarget
	target, err := h.assignmentService.GetAssignmentTarget(targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	// Проверяем, что задание принадлежит студенту
	if target.StudentID != studentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Получаем отправки студента
	submissions, err := h.submissionService.GetSubmissionsByAssignmentTarget(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get submissions"})
		return
	}

	// Получаем фидбэк учителя
	feedbacks, err := h.gradingService.GetFeedbacksByAssignmentTarget(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assignment":  target,
		"submissions": submissions,
		"feedbacks":   feedbacks,
	})
}

// POST /api/student/assignments/:id/submit - Отправить задание
func (h *StudentHandler) SubmitAssignment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	var request struct {
		Text     *string     `json:"text"`
		MediaIDs []uuid.UUID `json:"media_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Отправляем задание
	submission, err := h.submissionService.SubmitAssignment(targetID, studentID, request.Text, request.MediaIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submission": submission,
		"message":    "Assignment submitted successfully",
	})
}

// POST /api/student/assignments/:id/draft - Сохранить черновик
func (h *StudentHandler) SaveDraft(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	var request struct {
		Text     *string     `json:"text"`
		MediaIDs []uuid.UUID `json:"media_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Сохраняем черновик
	draft, err := h.submissionService.SaveDraft(targetID, studentID, request.Text, request.MediaIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draft":   draft,
		"message": "Draft saved successfully",
	})
}

// GET /api/student/assignments/:id/draft - Получить черновик
func (h *StudentHandler) GetDraft(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Получаем черновик
	draft, err := h.submissionService.GetDraft(targetID, studentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Draft not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"draft": draft,
	})
}

// GET /api/student/progress - Получить прогресс студента
func (h *StudentHandler) GetProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем все задания студента
	targets, err := h.assignmentService.GetAssignmentTargetsByStudent(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignments"})
		return
	}

	// Подсчитываем статистику
	var total, completed, graded, overdue int
	var totalScore float64
	var scoreCount int

	for _, target := range targets {
		total++
		switch target.Status {
		case models.AssignmentTargetStatusSubmitted, models.AssignmentTargetStatusGraded:
			completed++
			if target.Status == models.AssignmentTargetStatusGraded && target.Score != nil {
				totalScore += *target.Score
				scoreCount++
			}
		case models.AssignmentTargetStatusOverdue:
			overdue++
		}
	}

	var averageScore float64
	if scoreCount > 0 {
		averageScore = totalScore / float64(scoreCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":           total,
		"completed":       completed,
		"graded":          graded,
		"overdue":         overdue,
		"average_score":   averageScore,
		"completion_rate": float64(completed) / float64(total) * 100,
	})
}

// GET /api/student/notifications - Получить уведомления студента
func (h *StudentHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем параметры запроса
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Получаем уведомления
	notifications, err := h.notificationService.ListNotificationsByUser(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	// Применяем пагинацию
	start := offset
	end := offset + limit
	if start >= len(notifications) {
		notifications = []*models.Notification{}
	} else if end > len(notifications) {
		end = len(notifications)
	}

	result := notifications[start:end]

	c.JSON(http.StatusOK, gin.H{
		"notifications": result,
		"total":         len(notifications),
		"limit":         limit,
		"offset":        offset,
	})
}

// POST /api/student/notifications/:id/read - Отметить уведомление как прочитанное
func (h *StudentHandler) MarkNotificationAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	studentID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// Проверяем, что уведомление принадлежит пользователю
	notification, err := h.notificationService.GetNotification(notificationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if notification.UserID != studentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Отмечаем как прочитанное
	if err := h.notificationService.MarkAsRead(notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}
