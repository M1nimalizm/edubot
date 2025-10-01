package repository

import (
	"edubot/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TrialRequestRepository представляет репозиторий для работы с заявками на пробные занятия
type TrialRequestRepository struct {
	db *gorm.DB
}

// NewTrialRequestRepository создает новый репозиторий заявок на пробные занятия
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

// GetAll получает все заявки на пробные занятия
func (r *TrialRequestRepository) GetAll() ([]models.TrialRequest, error) {
	var requests []models.TrialRequest
	err := r.db.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// GetVisible получает только видимые заявки (не скрытые)
func (r *TrialRequestRepository) GetVisible() ([]models.TrialRequest, error) {
	var requests []models.TrialRequest
	err := r.db.Where("status != ?", "hidden").Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// GetByStatus получает заявки по статусу
func (r *TrialRequestRepository) GetByStatus(status string) ([]models.TrialRequest, error) {
	var requests []models.TrialRequest
	err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// Update обновляет заявку
func (r *TrialRequestRepository) Update(request *models.TrialRequest) error {
	return r.db.Save(request).Error
}

// Delete удаляет заявку
func (r *TrialRequestRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.TrialRequest{}, "id = ?", id).Error
}

// GetByTelegramID получает заявки по Telegram ID
func (r *TrialRequestRepository) GetByTelegramID(telegramID int64) ([]models.TrialRequest, error) {
	var requests []models.TrialRequest
	err := r.db.Where("telegram_id = ?", telegramID).Order("created_at DESC").Find(&requests).Error
	return requests, err
}
