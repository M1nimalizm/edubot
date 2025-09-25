package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

// MediaRepository интерфейс для работы с медиафайлами
type MediaRepository interface {
	Create(media *models.Media) error
	GetByID(id uuid.UUID) (*models.Media, error)
	GetByTelegramFileID(fileID string) (*models.Media, error)
	GetByOwnerID(ownerID uuid.UUID) ([]*models.Media, error)
	GetByEntity(entityType models.EntityType, entityID uuid.UUID) ([]*models.Media, error)
	GetByScope(scope models.MediaScope) ([]*models.Media, error)
	GetPublicMedia() ([]*models.Media, error)
	Update(media *models.Media) error
	Delete(id uuid.UUID) error
	GetRecentViews(mediaID uuid.UUID, limit int) ([]*models.MediaView, error)
	AddView(view *models.MediaView) error
	GetAccessList(mediaID uuid.UUID) ([]*models.MediaAccess, error)
	GrantAccess(access *models.MediaAccess) error
	RevokeAccess(mediaID, userID uuid.UUID) error
}

type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository создает новый репозиторий медиа
func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

// Create создает новый медиафайл
func (r *mediaRepository) Create(media *models.Media) error {
	if media.ID == uuid.Nil {
		media.ID = uuid.New()
	}
	return r.db.Create(media).Error
}

// GetByID получает медиафайл по ID
func (r *mediaRepository) GetByID(id uuid.UUID) (*models.Media, error) {
	var media models.Media
	err := r.db.Preload("Owner").First(&media, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// GetByTelegramFileID получает медиафайл по Telegram file ID
func (r *mediaRepository) GetByTelegramFileID(fileID string) (*models.Media, error) {
	var media models.Media
	err := r.db.Preload("Owner").First(&media, "telegram_file_id = ?", fileID).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// GetByOwnerID получает все медиафайлы владельца
func (r *mediaRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Media, error) {
	var media []*models.Media
	err := r.db.Preload("Owner").Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&media).Error
	return media, err
}

// GetByEntity получает медиафайлы по типу и ID сущности
func (r *mediaRepository) GetByEntity(entityType models.EntityType, entityID uuid.UUID) ([]*models.Media, error) {
	var media []*models.Media
	err := r.db.Preload("Owner").Where("entity_type = ? AND entity_id = ?", entityType, entityID).Order("created_at ASC").Find(&media).Error
	return media, err
}

// GetByScope получает медиафайлы по области видимости
func (r *mediaRepository) GetByScope(scope models.MediaScope) ([]*models.Media, error) {
	var media []*models.Media
	err := r.db.Preload("Owner").Where("scope = ?", scope).Order("created_at DESC").Find(&media).Error
	return media, err
}

// GetPublicMedia получает все публичные медиафайлы
func (r *mediaRepository) GetPublicMedia() ([]*models.Media, error) {
	return r.GetByScope(models.MediaScopePublic)
}

// Update обновляет медиафайл
func (r *mediaRepository) Update(media *models.Media) error {
	return r.db.Save(media).Error
}

// Delete удаляет медиафайл (мягкое удаление)
func (r *mediaRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Media{}, "id = ?", id).Error
}

// GetRecentViews получает последние просмотры медиафайла
func (r *mediaRepository) GetRecentViews(mediaID uuid.UUID, limit int) ([]*models.MediaView, error) {
	var views []*models.MediaView
	err := r.db.Preload("User").Where("media_id = ?", mediaID).Order("viewed_at DESC").Limit(limit).Find(&views).Error
	return views, err
}

// AddView добавляет запись о просмотре медиафайла
func (r *mediaRepository) AddView(view *models.MediaView) error {
	if view.ID == uuid.Nil {
		view.ID = uuid.New()
	}
	if view.ViewedAt.IsZero() {
		view.ViewedAt = time.Now()
	}
	return r.db.Create(view).Error
}

// GetAccessList получает список прав доступа к медиафайлу
func (r *mediaRepository) GetAccessList(mediaID uuid.UUID) ([]*models.MediaAccess, error) {
	var access []*models.MediaAccess
	err := r.db.Preload("User").Where("media_id = ?", mediaID).Find(&access).Error
	return access, err
}

// GrantAccess предоставляет доступ к медиафайлу
func (r *mediaRepository) GrantAccess(access *models.MediaAccess) error {
	if access.ID == uuid.Nil {
		access.ID = uuid.New()
	}
	if access.CreatedAt.IsZero() {
		access.CreatedAt = time.Now()
	}
	return r.db.Create(access).Error
}

// RevokeAccess отзывает доступ к медиафайлу
func (r *mediaRepository) RevokeAccess(mediaID, userID uuid.UUID) error {
	return r.db.Where("media_id = ? AND user_id = ?", mediaID, userID).Delete(&models.MediaAccess{}).Error
}
