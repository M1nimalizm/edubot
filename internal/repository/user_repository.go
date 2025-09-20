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

// TrialRequestRepository представляет репозиторий для работы с заявками на пробные занятия
type TrialRequestRepository struct {
	db *gorm.DB
}

// NewTrialRequestRepository создает новый репозиторий заявок
func NewTrialRequestRepository(db *gorm.DB) *TrialRequestRepository {
	return &TrialRequestRepository{db: db}
}

// Create создает новую заявку на пробное занятие
func (r *TrialRequestRepository) Create(request *models.TrialRequest) error {
	if request.ID == uuid.Nil {
		request.ID = uuid.New()
	}
	return r.db.Create(request).Error
}

// GetByID получает заявку по ID
func (r *TrialRequestRepository) GetByID(id uuid.UUID) (*models.TrialRequest, error) {
	var request models.TrialRequest
	err := r.db.First(&request, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// List получает список всех заявок
func (r *TrialRequestRepository) List() ([]models.TrialRequest, error) {
	var requests []models.TrialRequest
	err := r.db.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// ListByStatus получает заявки по статусу
func (r *TrialRequestRepository) ListByStatus(status string) ([]models.TrialRequest, error) {
	var requests []models.TrialRequest
	err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// Update обновляет заявку
func (r *TrialRequestRepository) Update(request *models.TrialRequest) error {
	return r.db.Save(request).Error
}

// UpdateStatus обновляет статус заявки
func (r *TrialRequestRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.TrialRequest{}).Where("id = ?", id).Update("status", status).Error
}

// Delete удаляет заявку
func (r *TrialRequestRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.TrialRequest{}, "id = ?", id).Error
}
