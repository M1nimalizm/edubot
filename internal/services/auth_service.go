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
	userRepo           repository.UserRepository
	trialRepo          *repository.TrialRequestRepository
	telegramBot        *telegram.Bot
	jwtSecret          string
	teacherTelegramID  int64
	teacherTelegramIDs map[int64]struct{}
	teacherPassword    string
}

// NewAuthService создает новый сервис авторизации
func NewAuthService(
	userRepo repository.UserRepository,
	trialRepo *repository.TrialRequestRepository,
	telegramBot *telegram.Bot,
	jwtSecret string,
	teacherTelegramID int64,
	teacherTelegramIDs []int64,
	teacherPassword string,
) *AuthService {
	idSet := make(map[int64]struct{})
	for _, id := range teacherTelegramIDs {
		idSet[id] = struct{}{}
	}
	return &AuthService{
		userRepo:           userRepo,
		trialRepo:          trialRepo,
		telegramBot:        telegramBot,
		jwtSecret:          jwtSecret,
		teacherTelegramID:  teacherTelegramID,
		teacherTelegramIDs: idSet,
		teacherPassword:    teacherPassword,
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
	User           *models.User `json:"user"`
	Token          string       `json:"token"`
	IsNewUser      bool         `json:"is_new_user"`
	Role           string       `json:"role"`
	AllowedTeacher bool         `json:"allowed_teacher"`
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
		if _, ok := s.teacherTelegramIDs[authData.ID]; ok || authData.ID == s.teacherTelegramID {
			user.Role = models.RoleTeacher
		}

		if err := s.userRepo.Create(user); err != nil {
			// Если параллельно уже создали этого пользователя — читаем его и считаем, что он не новый
			existing, gerr := s.userRepo.GetByTelegramID(authData.ID)
			if gerr != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			user = existing
			isNewUser = false
		} else {
			isNewUser = true
		}
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
		User:           user,
		Token:          token,
		IsNewUser:      isNewUser,
		Role:           string(user.Role),
		AllowedTeacher: s.IsTeacherTelegram(user.TelegramID),
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

	fmt.Printf("Sending trial request notification to teacher ID: %d\n", s.teacherTelegramID)
	if err := s.telegramBot.SendTrialRequestNotification(s.teacherTelegramID, requestData); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		fmt.Printf("Failed to send notification: %v\n", err)
	} else {
		fmt.Printf("Trial request notification sent successfully\n")
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

// Public helpers for handlers
func (s *AuthService) IsTeacherTelegram(telegramID int64) bool {
	if telegramID == s.teacherTelegramID {
		return true
	}
	_, ok := s.teacherTelegramIDs[telegramID]
	return ok
}

func (s *AuthService) ValidateTeacherPassword(password string) bool {
	if s.teacherPassword == "" {
		return false
	}
	return password == s.teacherPassword
}

func (s *AuthService) UpdateUser(user *models.User) error {
	return s.userRepo.Update(user)
}

func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	return s.generateJWT(user)
}

// AssignStudentParams параметры назначения ученика без участия ученика
type AssignStudentParams struct {
	UserID     *uuid.UUID
	TelegramID *int64
	Username   string
	Grade      *int
	Subjects   string
}

// AssignStudentToTeacher апгрейдит роль пользователя до student (teacher-driven)
func (s *AuthService) AssignStudentToTeacher(teacherID uuid.UUID, params AssignStudentParams) (*models.User, error) {
	var user *models.User
	var err error
	if params.UserID != nil {
		user, err = s.userRepo.GetByID(*params.UserID)
	} else if params.TelegramID != nil && *params.TelegramID != 0 {
		user, err = s.userRepo.GetByTelegramID(*params.TelegramID)
	} else if params.Username != "" {
		user, err = s.userRepo.GetByUsername(params.Username)
	} else {
		return nil, fmt.Errorf("no identifier provided")
	}
	if err != nil {
		return nil, err
	}

	// Обновление роли и учебных атрибутов
	user.Role = models.RoleStudent
	if params.Grade != nil {
		user.Grade = *params.Grade
	}
	if params.Subjects != "" {
		user.Subjects = params.Subjects
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Уведомление ученику
	if s.telegramBot != nil && user.TelegramID != 0 {
		s.telegramBot.SendMessage(user.TelegramID, "🎓 Вы назначены учеником. Откройте мини‑приложение для заданий.")
	}
	return user, nil
}

// SelectRole меняет роль пользователя и возвращает новый токен
func (s *AuthService) SelectRole(user *models.User, role models.UserRole) (string, error) {
	if role == models.RoleTeacher {
		if !s.IsTeacherTelegram(user.TelegramID) {
			return "", fmt.Errorf("not allowed to be teacher")
		}
	}
	user.Role = role
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}
	return s.generateJWT(user)
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

// LinkExistingStudentByUsername привязывает существующего пользователя к роли студента
func (s *AuthService) LinkExistingStudentByUsername(username string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	user.Role = models.RoleStudent
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// RegisterStudentByCode регистрирует ученика только по коду приглашения
func (s *AuthService) RegisterStudentByCode(inviteCode string) (*models.User, string, error) {
	// Проверяем код приглашения
	if inviteCode == "" {
		return nil, "", fmt.Errorf("invite code is required")
	}
	// Пытаемся найти уже созданного ученика с этим кодом (учитель создаёт заранее)
	if existing, err := s.userRepo.GetByInviteCode(inviteCode); err == nil && existing != nil {
		// Генерируем токен для существующего ученика
		token, err := s.generateJWT(existing)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate token: %w", err)
		}
		return existing, token, nil
	}

	// Если не нашли — создаём ученика (fallback)
	user := &models.User{
		ID:         uuid.New(),
		TelegramID: 0,
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

// ApproveTrialRequest одобряет заявку на пробный урок
func (s *AuthService) ApproveTrialRequest(requestID string) error {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return fmt.Errorf("invalid request ID: %w", err)
	}
	
	request, err := s.trialRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get trial request: %w", err)
	}

	request.Status = "approved"
	if err := s.trialRepo.Update(request); err != nil {
		return fmt.Errorf("failed to update trial request: %w", err)
	}

	// Отправляем уведомление заявителю (если есть telegram_id)
	if request.TelegramID != 0 && s.telegramBot != nil {
		message := fmt.Sprintf("🎉 Ваша заявка на пробный урок одобрена!\n\nМы свяжемся с вами в ближайшее время для согласования времени проведения занятия.")
		s.telegramBot.SendMessage(request.TelegramID, message)
	}

	return nil
}

// RejectTrialRequest отклоняет заявку на пробный урок
func (s *AuthService) RejectTrialRequest(requestID string) error {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return fmt.Errorf("invalid request ID: %w", err)
	}
	
	request, err := s.trialRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get trial request: %w", err)
	}

	request.Status = "rejected"
	if err := s.trialRepo.Update(request); err != nil {
		return fmt.Errorf("failed to update trial request: %w", err)
	}

	// Отправляем уведомление заявителю (если есть telegram_id)
	if request.TelegramID != 0 && s.telegramBot != nil {
		message := fmt.Sprintf("К сожалению, ваша заявка на пробный урок отклонена.\n\nЕсли у вас есть вопросы, пожалуйста, свяжитесь с преподавателем.")
		s.telegramBot.SendMessage(request.TelegramID, message)
	}

	return nil
}

// SearchUsers ищет пользователей по запросу (только гости)
func (s *AuthService) SearchUsers(query string) ([]models.User, error) {
	// Ищем только пользователей с ролью "guest"
	users, err := s.userRepo.SearchByQuery(query, "guest")
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	
	return users, nil
}
