package repository

import (
	"edubot/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HomepageMediaRepository interface {
	Create(media *models.HomepageMedia) error
	GetByID(id uuid.UUID) (*models.HomepageMedia, error)
	GetByType(mediaType models.HomepageMediaType) (*models.HomepageMedia, error)
	GetActiveByType(mediaType models.HomepageMediaType) (*models.HomepageMedia, error)
	List() ([]*models.HomepageMedia, error)
	Update(media *models.HomepageMedia) error
	Delete(id uuid.UUID) error
	SetActive(mediaType models.HomepageMediaType, mediaID uuid.UUID) error
}

type homepageMediaRepository struct {
	db *gorm.DB
}

func NewHomepageMediaRepository(db *gorm.DB) HomepageMediaRepository {
	return &homepageMediaRepository{db: db}
}

func (r *homepageMediaRepository) Create(media *models.HomepageMedia) error {
	return r.db.Create(media).Error
}

func (r *homepageMediaRepository) GetByID(id uuid.UUID) (*models.HomepageMedia, error) {
	var media models.HomepageMedia
	err := r.db.Where("id = ?", id).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *homepageMediaRepository) GetByType(mediaType models.HomepageMediaType) (*models.HomepageMedia, error) {
	var media models.HomepageMedia
	err := r.db.Where("type = ?", mediaType).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *homepageMediaRepository) GetActiveByType(mediaType models.HomepageMediaType) (*models.HomepageMedia, error) {
	var media models.HomepageMedia
	err := r.db.Where("type = ? AND is_active = ?", mediaType, true).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *homepageMediaRepository) List() ([]*models.HomepageMedia, error) {
	var media []*models.HomepageMedia
	err := r.db.Order("created_at DESC").Find(&media).Error
	return media, err
}

func (r *homepageMediaRepository) Update(media *models.HomepageMedia) error {
	return r.db.Save(media).Error
}

func (r *homepageMediaRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.HomepageMedia{}, id).Error
}

func (r *homepageMediaRepository) SetActive(mediaType models.HomepageMediaType, mediaID uuid.UUID) error {
	// Деактивируем все файлы этого типа
	err := r.db.Model(&models.HomepageMedia{}).Where("type = ?", mediaType).Update("is_active", false).Error
	if err != nil {
		return err
	}
	
	// Активируем выбранный файл
	return r.db.Model(&models.HomepageMedia{}).Where("id = ?", mediaID).Update("is_active", true).Error
}
