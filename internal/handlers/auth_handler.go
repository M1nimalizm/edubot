package handlers

import (
	"net/http"
	"strconv"

	"edubot/internal/models"
	"edubot/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// RegisterStudentByCodeRequest представляет запрос регистрации ученика только по коду
type RegisterStudentByCodeRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// TeacherLoginRequest запрос для входа учителя (после Telegram-авторизации)
type TeacherLoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// SelectRoleRequest запрос на выбор роли (учитель/ученик)
type SelectRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=teacher student"`
}

// CreateStudentByTeacherRequest запрос на создание ученика преподавателем
type CreateStudentByTeacherRequest struct {
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name"`
	Grade      int    `json:"grade" binding:"required,min=1,max=11"`
	Subjects   string `json:"subjects"`
	Phone      string `json:"phone"`
	Username   string `json:"username"`
	TelegramID int64  `json:"telegram_id"`
}

// AssignStudentRequest — teacher-driven назначение ученика без кода
type AssignStudentRequest struct {
	UserID     *uuid.UUID `json:"user_id"`
	TelegramID *int64     `json:"telegram_id"`
	Username   string     `json:"username"`
	Grade      *int       `json:"grade"`
	Subjects   string     `json:"subjects"`
}

// TrialRequestRequest представляет запрос на пробное занятие
type TrialRequestRequest struct {
	Name         string `json:"name" binding:"required"`
	Grade        int    `json:"grade" binding:"required,min=10,max=11"`
	Subject      string `json:"subject" binding:"required"`
	Level        int    `json:"level" binding:"required,min=1,max=5"`
	Comment      string `json:"comment"`
	ContactType  string `json:"contact_type" binding:"required"`
	ContactValue string `json:"contact_value" binding:"required"`
}

// SearchUsersRequest запрос для поиска пользователей
type SearchUsersRequest struct {
	Query string `json:"query" binding:"required"`
}

// SearchUsersResponse ответ поиска пользователей
type SearchUsersResponse struct {
	Users []models.User `json:"users"`
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

	// Ставим jwt в cookie для удобства переходов по HTML
	if result != nil && result.Token != "" {
		c.SetCookie("jwt", result.Token, 3600*24, "/", "", false, true)
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

	// Для публичных заявок telegram_id не обязателен
	trialRequest := &models.TrialRequest{
		Name:         req.Name,
		Grade:        req.Grade,
		Subject:      req.Subject,
		Level:        req.Level,
		Comment:      req.Comment,
		ContactType:  req.ContactType,
		ContactValue: req.ContactValue,
		TelegramID:   0, // Публичные заявки без telegram_id
	}

	if err := h.authService.SubmitTrialRequest(trialRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trial request submitted successfully"})
}

// GenerateInviteCode генерирует код приглашения
func (h *AuthHandler) GenerateInviteCode(c *gin.Context) {
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

// GetStudents получает список учеников для учителя
func (h *AuthHandler) GetStudents(c *gin.Context) {
	students, err := h.authService.GetStudents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// CreateStudentByTeacher создает ученика и возвращает код приглашения
func (h *AuthHandler) CreateStudentByTeacher(c *gin.Context) {
	var req CreateStudentByTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Если указан username без остальных полей — пытаемся привязать существующего
	if req.Username != "" && req.FirstName == "" && req.Grade == 0 {
		user, err := h.authService.LinkExistingStudentByUsername(req.Username)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": user, "invite_code": user.InviteCode})
		return
	}

	user, code, err := h.authService.CreateStudentByTeacher(
		req.FirstName,
		req.LastName,
		req.Grade,
		req.Subjects,
		req.Phone,
		req.Username,
		req.TelegramID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"invite_code": code,
	})
}

// AssignStudentToTeacher назначает гостя учеником (teacher-only)
func (h *AuthHandler) AssignStudentToTeacher(c *gin.Context) {
	// Проверяем роль
	roleVal, exists := c.Get("user_role")
	if !exists || roleVal != models.RoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	userVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	teacherID, ok := userVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	var req AssignStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.AssignStudentToTeacher(teacherID, services.AssignStudentParams{
		UserID:     req.UserID,
		TelegramID: req.TelegramID,
		Username:   req.Username,
		Grade:      req.Grade,
		Subjects:   req.Subjects,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// RegisterStudentByCode регистрирует ученика только по коду приглашения
func (h *AuthHandler) RegisterStudentByCode(c *gin.Context) {
	var req RegisterStudentByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем временного пользователя и регистрируем его как ученика
	user, token, err := h.authService.RegisterStudentByCode(req.InviteCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student registered successfully",
		"user":    user,
		"token":   token,
	})
}

// TeacherLogin после Telegram-авторизации повышает роль до teacher при верном пароле и валидном Telegram ID
func (h *AuthHandler) TeacherLogin(c *gin.Context) {
	var req TeacherLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Достаём пользователя из контекста (AuthMiddleware должен быть активен)
	userAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user := userAny.(*models.User)

	// Проверяем право на teacher: Telegram ID должен быть в белом списке
	if !h.authService.IsTeacherTelegram(user.TelegramID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Проверяем пароль
	if !h.authService.ValidateTeacherPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Повышаем роль до teacher (единый учёт у всех разрешённых ID)
	user.Role = models.RoleTeacher
	if err := h.authService.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Выдаём новый токен с ролью teacher
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Teacher login successful", "token": token, "user": user})
}

// SelectRole устанавливает роль пользователя (для whitelisted учителей)
func (h *AuthHandler) SelectRole(c *gin.Context) {
	var req SelectRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userAny.(*models.User)

	var role models.UserRole
	if req.Role == "teacher" {
		role = models.RoleTeacher
	} else {
		role = models.RoleStudent
	}

	token, err := h.authService.SelectRole(user, role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "role": role})
}

// ApproveTrialRequest одобряет заявку на пробный урок
func (h *AuthHandler) ApproveTrialRequest(c *gin.Context) {
	// Проверяем роль
	roleVal, exists := c.Get("user_role")
	if !exists || roleVal != models.RoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request ID is required"})
		return
	}

	err := h.authService.ApproveTrialRequest(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trial request approved successfully"})
}

// RejectTrialRequest отклоняет заявку на пробный урок
func (h *AuthHandler) RejectTrialRequest(c *gin.Context) {
	// Проверяем роль
	roleVal, exists := c.Get("user_role")
	if !exists || roleVal != models.RoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request ID is required"})
		return
	}

	err := h.authService.RejectTrialRequest(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trial request rejected successfully"})
}

// SearchUsers ищет пользователей по запросу (только гости)
func (h *AuthHandler) SearchUsers(c *gin.Context) {
	var req SearchUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := h.authService.SearchUsers(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SearchUsersResponse{Users: users})
}

// HideTrialRequest скрывает заявку на пробный урок (не удаляет из БД)
func (h *AuthHandler) HideTrialRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request ID is required"})
		return
	}

	err := h.authService.HideTrialRequest(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trial request hidden successfully"})
}
