package repository

import (
	"fmt"
	"time"

	"edubot/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByTelegramID(telegramID int64) (*models.User, error)
	GetByInviteCode(code string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	ListStudents() ([]models.User, error)
	ListByRole(role models.UserRole) ([]models.User, error)
	GenerateInviteCode() (string, error)
	SearchByQuery(query string, role string) ([]models.User, error)
}

// userRepository реализация репозитория пользователей
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create создает нового пользователя
func (r *userRepository) Create(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
    return r.db.Create(user).Error
}

// GetByID получает пользователя по ID
func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByTelegramID получает пользователя по Telegram ID
func (r *userRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	err := r.db.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByInviteCode получает пользователя по коду приглашения
func (r *userRepository) GetByInviteCode(code string) (*models.User, error) {
	var user models.User
	err := r.db.Where("invite_code = ?", code).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername получает пользователя по Telegram username
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
    var user models.User
    err := r.db.Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Update обновляет пользователя
func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete удаляет пользователя
func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// ListStudents получает список всех учеников
func (r *userRepository) ListStudents() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ?", models.RoleStudent).Find(&users).Error
	return users, err
}

// ListByRole получает пользователей по роли
func (r *userRepository) ListByRole(role models.UserRole) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ?", role).Find(&users).Error
	return users, err
}

// GenerateInviteCode генерирует уникальный код приглашения
func (r *userRepository) GenerateInviteCode() (string, error) {
	for {
		code := fmt.Sprintf("EDU%d", time.Now().Unix()%100000)

		var count int64
		r.db.Model(&models.User{}).Where("invite_code = ?", code).Count(&count)

		if count == 0 {
			return code, nil
		}
	}
}

// SearchByQuery ищет пользователей по запросу с фильтром по роли
func (r *userRepository) SearchByQuery(query string, role string) ([]models.User, error) {
	var users []models.User
	
	// Поиск по имени, фамилии, username или telegram_id
	err := r.db.Where("role = ? AND (first_name ILIKE ? OR last_name ILIKE ? OR username ILIKE ? OR telegram_id::text ILIKE ?)", 
		role, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Order("created_at DESC").
		Find(&users).Error
	
	return users, err
}
