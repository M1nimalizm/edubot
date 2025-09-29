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

type TeacherInboxHandler struct {
	gradingService      services.GradingService
	assignmentService   services.AssignmentService
	submissionService   services.SubmissionService
	chatService         services.ChatService
	notificationService services.NotificationService
}

func NewTeacherInboxHandler(
	gradingService services.GradingService,
	assignmentService services.AssignmentService,
	submissionService services.SubmissionService,
	chatService services.ChatService,
	notificationService services.NotificationService,
) *TeacherInboxHandler {
	return &TeacherInboxHandler{
		gradingService:      gradingService,
		assignmentService:   assignmentService,
		submissionService:   submissionService,
		chatService:         chatService,
		notificationService: notificationService,
	}
}

// GET /api/teacher/inbox - Получить inbox учителя
func (h *TeacherInboxHandler) GetInbox(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем параметры запроса
	status := c.Query("status") // pending, graded, all
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

	var targets []*models.AssignmentTarget

	switch status {
	case "pending":
		targets, err = h.gradingService.GetPendingGrading(teacherID)
	case "graded":
		targets, err = h.gradingService.GetGradedAssignments(teacherID)
	default:
		// Получаем все задания учителя
		assignments, err := h.assignmentService.ListAssignmentsByTeacher(teacherID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignments"})
			return
		}

		// Получаем все таргеты для этих заданий
		for _, assignment := range assignments {
			assignmentTargets, err := h.assignmentService.GetAssignmentTargetsByAssignment(assignment.ID)
			if err == nil {
				targets = append(targets, assignmentTargets...)
			}
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignments"})
		return
	}

	// Применяем пагинацию
	start := offset
	end := offset + limit
	if start >= len(targets) {
		targets = []*models.AssignmentTarget{}
	} else if end > len(targets) {
		end = len(targets)
	}

	result := targets[start:end]

	c.JSON(http.StatusOK, gin.H{
		"assignments": result,
		"total":       len(targets),
		"limit":       limit,
		"offset":      offset,
	})
}

// GET /api/teacher/inbox/:id - Получить детали задания для оценки
func (h *TeacherInboxHandler) GetAssignmentForGrading(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
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

	// Проверяем, что задание принадлежит учителю
	if target.Assignment.TeacherID != teacherID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Получаем отправки студента
	submissions, err := h.submissionService.GetSubmissionsByAssignmentTarget(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get submissions"})
		return
	}

	// Получаем предыдущие фидбэки
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

// POST /api/teacher/inbox/:id/grade - Оценить задание
func (h *TeacherInboxHandler) GradeAssignment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
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
		Score    *float64    `json:"score"`
		Text     string      `json:"text"`
		MediaIDs []uuid.UUID `json:"media_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Оцениваем задание
	feedback, err := h.gradingService.GradeAssignment(targetID, teacherID, request.Score, request.Text, request.MediaIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feedback": feedback,
		"message":  "Assignment graded successfully",
	})
}

// GET /api/teacher/assignments - Получить все задания учителя
func (h *TeacherInboxHandler) GetAssignments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем параметры запроса
	groupIDStr := c.Query("group_id")
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

	var assignments []*models.Assignment

	if groupIDStr != "" {
		groupID, err := uuid.Parse(groupIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}
		assignments, err = h.assignmentService.ListAssignmentsByGroup(groupID)
	} else {
		assignments, err = h.assignmentService.ListAssignmentsByTeacher(teacherID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignments"})
		return
	}

	// Применяем пагинацию
	start := offset
	end := offset + limit
	if start >= len(assignments) {
		assignments = []*models.Assignment{}
	} else if end > len(assignments) {
		end = len(assignments)
	}

	result := assignments[start:end]

	c.JSON(http.StatusOK, gin.H{
		"assignments": result,
		"total":       len(assignments),
		"limit":       limit,
		"offset":      offset,
	})
}

// POST /api/teacher/assignments - Создать новое задание
func (h *TeacherInboxHandler) CreateAssignment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request struct {
		GroupID     *uuid.UUID `json:"group_id"`
		StudentID   *uuid.UUID `json:"student_id"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Subject     string     `json:"subject"`
		Grade       int        `json:"grade"`
		Level       int        `json:"level"`
		DueDate     string     `json:"due_date"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Парсим дату дедлайна
	dueDate, err := time.Parse(time.RFC3339, request.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due date format"})
		return
	}

	var assignment *models.Assignment

	if request.GroupID != nil {
		// Создаем групповое задание
		assignment, err = h.assignmentService.CreateGroupAssignment(
			teacherID,
			*request.GroupID,
			request.Title,
			request.Description,
			request.Subject,
			request.Grade,
			request.Level,
			dueDate,
		)
	} else {
		// Создаем индивидуальное задание
		assignment = &models.Assignment{
			Title:       request.Title,
			Description: request.Description,
			Subject:     request.Subject,
			Grade:       request.Grade,
			Level:       request.Level,
			TeacherID:   teacherID,
			StudentID:   request.StudentID,
			DueDate:     dueDate,
			Status:      "active",
			CreatedBy:   teacherID,
		}

		err = h.assignmentService.CreateAssignment(assignment)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assignment": assignment,
		"message":    "Assignment created successfully",
	})
}

// GET /api/teacher/statistics - Получить статистику учителя
func (h *TeacherInboxHandler) GetStatistics(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем все задания учителя
	assignments, err := h.assignmentService.ListAssignmentsByTeacher(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignments"})
		return
	}

	// Подсчитываем статистику
	var totalAssignments, pendingGrading, gradedAssignments int
	var totalScore float64
	var scoreCount int

	for _, assignment := range assignments {
		totalAssignments++

		// Получаем таргеты для этого задания
		targets, err := h.assignmentService.GetAssignmentTargetsByAssignment(assignment.ID)
		if err != nil {
			continue
		}

		for _, target := range targets {
			switch target.Status {
			case models.AssignmentTargetStatusSubmitted:
				pendingGrading++
			case models.AssignmentTargetStatusGraded:
				gradedAssignments++
				if target.Score != nil {
					totalScore += *target.Score
					scoreCount++
				}
			}
		}
	}

	var averageScore float64
	if scoreCount > 0 {
		averageScore = totalScore / float64(scoreCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_assignments":  totalAssignments,
		"pending_grading":    pendingGrading,
		"graded_assignments": gradedAssignments,
		"average_score":      averageScore,
		"completion_rate":    float64(gradedAssignments) / float64(totalAssignments) * 100,
	})
}

// GET /api/teacher/notifications - Получить уведомления учителя
func (h *TeacherInboxHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
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
	notifications, err := h.notificationService.ListNotificationsByUser(teacherID)
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

// POST /api/teacher/notifications/:id/read - Отметить уведомление как прочитанное
func (h *TeacherInboxHandler) MarkNotificationAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	teacherID, ok := userID.(uuid.UUID)
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

	if notification.UserID != teacherID {
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
