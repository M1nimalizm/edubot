package services

import (
	"fmt"
	"time"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthService представляет сервис авторизации
type AuthService struct {
	userRepo          *repository.UserRepository
	trialRepo         *repository.TrialRequestRepository
	telegramBot       *telegram.Bot
	jwtSecret         string
	teacherTelegramID int64
}

// NewAuthService создает новый сервис авторизации
func NewAuthService(
	userRepo *repository.UserRepository,
	trialRepo *repository.TrialRequestRepository,
	telegramBot *telegram.Bot,
	jwtSecret string,
	teacherTelegramID int64,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		trialRepo:         trialRepo,
		telegramBot:       telegramBot,
		jwtSecret:         jwtSecret,
		teacherTelegramID: teacherTelegramID,
	}
}

// TelegramAuthData представляет данные авторизации из Telegram
type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// AuthResult представляет результат авторизации
type AuthResult struct {
	User      *models.User `json:"user"`
	Token     string       `json:"token"`
	IsNewUser bool         `json:"is_new_user"`
	Role      string       `json:"role"`
}

// AuthenticateWithTelegram авторизует пользователя через Telegram
func (s *AuthService) AuthenticateWithTelegram(authData *TelegramAuthData) (*AuthResult, error) {
	// Проверяем подпись данных (в реальном приложении)
	// if !s.validateTelegramAuth(authData) {
	//     return nil, fmt.Errorf("invalid telegram auth data")
	// }

	// Ищем существующего пользователя
	user, err := s.userRepo.GetByTelegramID(authData.ID)
	isNewUser := false

	if err != nil {
		// Пользователь не найден, создаем нового
		user = &models.User{
			TelegramID: authData.ID,
			FirstName:  authData.FirstName,
			LastName:   authData.LastName,
			Username:   authData.Username,
			Role:       models.RoleGuest, // По умолчанию гость
			InviteCode: nil,              // Пустой invite code
		}

		// Проверяем, является ли пользователь преподавателем
		if authData.ID == s.teacherTelegramID {
			user.Role = models.RoleTeacher
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		isNewUser = true
	} else {
		// Обновляем данные существующего пользователя
		user.FirstName = authData.FirstName
		user.LastName = authData.LastName
		user.Username = authData.Username
		if err := s.userRepo.Update(user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Генерируем JWT токен
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResult{
		User:      user,
		Token:     token,
		IsNewUser: isNewUser,
		Role:      string(user.Role),
	}, nil
}

// RegisterStudent регистрирует ученика по коду приглашения
func (s *AuthService) RegisterStudent(telegramID int64, inviteCode string, phone string, grade int, subjects string) error {
	// Получаем пользователя
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем код приглашения (в реальном приложении нужно проверить валидность кода)
	if inviteCode == "" {
		return fmt.Errorf("invite code is required")
	}

	// Обновляем данные пользователя
	user.Role = models.RoleStudent
	user.Phone = phone
	user.Grade = grade
	user.Subjects = subjects
	user.InviteCode = &inviteCode

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to register student: %w", err)
	}

	return nil
}

// SubmitTrialRequest создает заявку на пробное занятие
func (s *AuthService) SubmitTrialRequest(request *models.TrialRequest) error {
	// Создаем заявку
	if err := s.trialRepo.Create(request); err != nil {
		return fmt.Errorf("failed to create trial request: %w", err)
	}

	// Отправляем уведомление преподавателю
	requestData := map[string]interface{}{
		"name":          request.Name,
		"grade":         request.Grade,
		"subject":       request.Subject,
		"level":         request.Level,
		"contact_type":  request.ContactType,
		"contact_value": request.ContactValue,
		"comment":       request.Comment,
		"created_at":    request.CreatedAt.Format("02.01.2006 15:04"),
	}

	if err := s.telegramBot.SendTrialRequestNotification(s.teacherTelegramID, requestData); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		fmt.Printf("Failed to send notification: %v\n", err)
	}

	return nil
}

// GenerateInviteCode генерирует код приглашения для преподавателя
func (s *AuthService) GenerateInviteCode() (string, error) {
	return s.userRepo.GenerateInviteCode()
}

// ValidateToken валидирует JWT токен
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID in token: %w", err)
		}

		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// generateJWT генерирует JWT токен для пользователя
func (s *AuthService) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     user.ID.String(),
		"telegram_id": user.TelegramID,
		"role":        user.Role,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GetTrialRequests получает все заявки на пробные занятия
func (s *AuthService) GetTrialRequests() ([]models.TrialRequest, error) {
	return s.trialRepo.GetAll()
}

// GetStats получает статистику для панели управления
func (s *AuthService) GetStats() (map[string]interface{}, error) {
	// Получаем количество учеников
	students, err := s.userRepo.ListByRole(models.RoleStudent)
	if err != nil {
		return nil, err
	}

	// Получаем количество заявок
	requests, err := s.trialRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Подсчитываем новые заявки (статус pending)
	pendingCount := 0
	for _, req := range requests {
		if req.Status == "pending" {
			pendingCount++
		}
	}

	return map[string]interface{}{
		"total_students":    len(students),
		"pending_requests":  pendingCount,
		"total_requests":    len(requests),
		"total_assignments": 0, // TODO: реализовать когда будут задания
		"total_content":     0, // TODO: реализовать когда будет контент
	}, nil
}

// GetStudents получает список всех учеников
func (s *AuthService) GetStudents() ([]models.User, error) {
	return s.userRepo.ListByRole("student")
}

// RegisterStudentByCode регистрирует ученика только по коду приглашения
func (s *AuthService) RegisterStudentByCode(inviteCode string) (*models.User, string, error) {
	// Проверяем код приглашения
	if inviteCode == "" {
		return nil, "", fmt.Errorf("invite code is required")
	}

	// Создаем нового пользователя-ученика
	user := &models.User{
		ID:         uuid.New(),
		TelegramID: 0, // Временно 0, будет заполнено позже
		Role:       models.RoleStudent,
		InviteCode: &inviteCode,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем JWT токен
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// CreateStudentByTeacher создает ученика от имени преподавателя и выдает код приглашения
func (s *AuthService) CreateStudentByTeacher(firstName string, lastName string, grade int, subjects string, phone string, username string, telegramID int64) (*models.User, string, error) {
    // Сгенерировать уникальный инвайт-код
    code, err := s.userRepo.GenerateInviteCode()
    if err != nil {
        return nil, "", fmt.Errorf("failed to generate invite code: %w", err)
    }

    // Сформировать пользователя
    user := &models.User{
        ID:         uuid.New(),
        TelegramID: telegramID, // может быть 0, если неизвестен
        Username:   username,
        FirstName:  firstName,
        LastName:   lastName,
        Role:       models.RoleStudent,
        Phone:      phone,
        Grade:      grade,
        Subjects:   subjects,
        InviteCode: &code,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }

    if err := s.userRepo.Create(user); err != nil {
        return nil, "", fmt.Errorf("failed to create student: %w", err)
    }

    return user, code, nil
}

// validateTelegramAuth валидирует данные авторизации Telegram (упрощенная версия)
func (s *AuthService) validateTelegramAuth(authData *TelegramAuthData) bool {
	// В реальном приложении здесь должна быть проверка подписи
	// с использованием секретного ключа бота
	return true
}
