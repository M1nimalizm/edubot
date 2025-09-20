package repository

import (
	"fmt"
	"time"

	"edubot/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository представляет репозиторий для работы с пользователями
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create создает нового пользователя
func (r *UserRepository) Create(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return r.db.Create(user).Error
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByTelegramID получает пользователя по Telegram ID
func (r *UserRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	err := r.db.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByInviteCode получает пользователя по коду приглашения
func (r *UserRepository) GetByInviteCode(code string) (*models.User, error) {
	var user models.User
	err := r.db.Where("invite_code = ?", code).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update обновляет пользователя
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// ListStudents получает список всех учеников
func (r *UserRepository) ListStudents() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ?", models.RoleStudent).Find(&users).Error
	return users, err
}

// ListByRole получает пользователей по роли
func (r *UserRepository) ListByRole(role models.UserRole) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ?", role).Find(&users).Error
	return users, err
}

// GenerateInviteCode генерирует уникальный код приглашения
func (r *UserRepository) GenerateInviteCode() (string, error) {
	for {
		code := fmt.Sprintf("EDU%d", time.Now().Unix()%100000)

		var count int64
		r.db.Model(&models.User{}).Where("invite_code = ?", code).Count(&count)

		if count == 0 {
			return code, nil
		}
	}
}

