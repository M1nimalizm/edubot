package handlers

import (
	"net/http"
	"strconv"

	"edubot/internal/models"
	"edubot/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler представляет обработчик авторизации
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler создает новый обработчик авторизации
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// TelegramAuthRequest представляет запрос авторизации через Telegram
type TelegramAuthRequest struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// RegisterStudentRequest представляет запрос регистрации ученика
type RegisterStudentRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Grade      int    `json:"grade" binding:"required,min=10,max=11"`
	Subjects   string `json:"subjects" binding:"required"`
}

// TrialRequestRequest представляет запрос на пробное занятие
type TrialRequestRequest struct {
	Name    string `json:"name" binding:"required"`
	Grade   int    `json:"grade" binding:"required,min=10,max=11"`
	Subject string `json:"subject" binding:"required"`
	Level   int    `json:"level" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
	Phone   string `json:"phone" binding:"required"`
}

// TelegramAuth авторизует пользователя через Telegram
func (h *AuthHandler) TelegramAuth(c *gin.Context) {
	var req TelegramAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authData := &services.TelegramAuthData{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		PhotoURL:  req.PhotoURL,
		AuthDate:  req.AuthDate,
		Hash:      req.Hash,
	}

	result, err := h.authService.AuthenticateWithTelegram(authData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RegisterStudent регистрирует ученика по коду приглашения
func (h *AuthHandler) RegisterStudent(c *gin.Context) {
	var req RegisterStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем Telegram ID из контекста (должен быть установлен middleware)
	telegramIDStr, exists := c.Get("telegram_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "telegram_id not found"})
		return
	}

	telegramID, err := strconv.ParseInt(telegramIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid telegram_id"})
		return
	}

	if err := h.authService.RegisterStudent(telegramID, req.InviteCode, req.Phone, req.Grade, req.Subjects); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student registered successfully"})
}

// SubmitTrialRequest создает заявку на пробное занятие
func (h *AuthHandler) SubmitTrialRequest(c *gin.Context) {
	var req TrialRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем Telegram ID из контекста
	telegramIDStr, exists := c.Get("telegram_id")
	var telegramID int64
	if exists {
		if id, err := strconv.ParseInt(telegramIDStr.(string), 10, 64); err == nil {
			telegramID = id
		}
	}

	trialRequest := &models.TrialRequest{
		Name:       req.Name,
		Grade:      req.Grade,
		Subject:    req.Subject,
		Level:      req.Level,
		Comment:    req.Comment,
		Phone:      req.Phone,
		TelegramID: telegramID,
	}

	if err := h.authService.SubmitTrialRequest(trialRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trial request submitted successfully"})
}

// GenerateInviteCode генерирует код приглашения (только для преподавателя)
func (h *AuthHandler) GenerateInviteCode(c *gin.Context) {
	// Проверяем роль пользователя
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	code, err := h.authService.GenerateInviteCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invite_code": code})
}

// GetProfile получает профиль пользователя
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetTrialRequests получает заявки на пробные занятия
func (h *AuthHandler) GetTrialRequests(c *gin.Context) {
	requests, err := h.authService.GetTrialRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

// GetStats получает статистику для панели управления
func (h *AuthHandler) GetStats(c *gin.Context) {
	stats, err := h.authService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
