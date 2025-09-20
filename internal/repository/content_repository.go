package repository

import (
	"edubot/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentRepository представляет репозиторий для работы с образовательным контентом
type ContentRepository struct {
	db *gorm.DB
}

// NewContentRepository создает новый репозиторий контента
func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

// Create создает новый контент
func (r *ContentRepository) Create(content *models.Content) error {
	if content.ID == uuid.Nil {
		content.ID = uuid.New()
	}
	return r.db.Create(content).Error
}

// GetByID получает контент по ID
func (r *ContentRepository) GetByID(id uuid.UUID) (*models.Content, error) {
	var content models.Content
	err := r.db.Preload("Creator").Preload("Attachments").
		First(&content, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// List получает список всего контента
func (r *ContentRepository) List() ([]models.Content, error) {
	var contents []models.Content
	err := r.db.Where("is_public = ?", true).Preload("Creator").Preload("Attachments").
		Order("created_at DESC").Find(&contents).Error
	return contents, err
}

// ListByType получает контент по типу
func (r *ContentRepository) ListByType(contentType string) ([]models.Content, error) {
	var contents []models.Content
	err := r.db.Where("type = ? AND is_public = ?", contentType, true).
		Preload("Creator").Preload("Attachments").
		Order("created_at DESC").Find(&contents).Error
	return contents, err
}

// ListByCategory получает контент по категории
func (r *ContentRepository) ListByCategory(category string) ([]models.Content, error) {
	var contents []models.Content
	err := r.db.Where("category = ? AND is_public = ?", category, true).
		Preload("Creator").Preload("Attachments").
		Order("created_at DESC").Find(&contents).Error
	return contents, err
}

// ListByCreator получает контент по создателю
func (r *ContentRepository) ListByCreator(creatorID uuid.UUID) ([]models.Content, error) {
	var contents []models.Content
	err := r.db.Where("created_by = ?", creatorID).
		Preload("Creator").Preload("Attachments").
		Order("created_at DESC").Find(&contents).Error
	return contents, err
}

// Search выполняет поиск контента
func (r *ContentRepository) Search(query string) ([]models.Content, error) {
	var contents []models.Content
	err := r.db.Where("is_public = ? AND (title ILIKE ? OR description ILIKE ? OR tags ILIKE ?)",
		true, "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Preload("Creator").Preload("Attachments").
		Order("created_at DESC").Find(&contents).Error
	return contents, err
}

// Update обновляет контент
func (r *ContentRepository) Update(content *models.Content) error {
	return r.db.Save(content).Error
}

// Delete удаляет контент
func (r *ContentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Content{}, "id = ?", id).Error
}

// MarkAsViewed отмечает контент как просмотренный пользователем
func (r *ContentRepository) MarkAsViewed(contentID, userID uuid.UUID) error {
	// Проверяем, не просматривал ли пользователь этот контент ранее
	var count int64
	r.db.Model(&models.ContentView{}).Where("content_id = ? AND user_id = ?", contentID, userID).Count(&count)

	if count == 0 {
		view := models.ContentView{
			ID:        uuid.New(),
			ContentID: contentID,
			UserID:    userID,
		}
		return r.db.Create(&view).Error
	}

	return nil
}

// GetViewedContent получает просмотренный пользователем контент
func (r *ContentRepository) GetViewedContent(userID uuid.UUID) ([]models.Content, error) {
	var contents []models.Content
	err := r.db.Joins("JOIN content_views ON contents.id = content_views.content_id").
		Where("content_views.user_id = ?", userID).
		Preload("Creator").Preload("Attachments").
		Order("content_views.viewed_at DESC").Find(&contents).Error
	return contents, err
}

// AttachmentRepository представляет репозиторий для работы с файлами
type AttachmentRepository struct {
	db *gorm.DB
}

// NewAttachmentRepository создает новый репозиторий файлов
func NewAttachmentRepository(db *gorm.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create создает новую запись о файле
func (r *AttachmentRepository) Create(attachment *models.Attachment) error {
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}
	return r.db.Create(attachment).Error
}

// GetByID получает файл по ID
func (r *AttachmentRepository) GetByID(id uuid.UUID) (*models.Attachment, error) {
	var attachment models.Attachment
	err := r.db.First(&attachment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// ListByAssignment получает файлы задания
func (r *AttachmentRepository) ListByAssignment(assignmentID uuid.UUID) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("assignment_id = ?", assignmentID).Find(&attachments).Error
	return attachments, err
}

// ListBySubmission получает файлы решения
func (r *AttachmentRepository) ListBySubmission(submissionID uuid.UUID) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("submission_id = ?", submissionID).Find(&attachments).Error
	return attachments, err
}

// ListByContent получает файлы контента
func (r *AttachmentRepository) ListByContent(contentID uuid.UUID) ([]models.Attachment, error) {
	var attachments []models.Attachment
	err := r.db.Where("content_id = ?", contentID).Find(&attachments).Error
	return attachments, err
}

// Update обновляет информацию о файле
func (r *AttachmentRepository) Update(attachment *models.Attachment) error {
	return r.db.Save(attachment).Error
}

// Delete удаляет запись о файле
func (r *AttachmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Attachment{}, "id = ?", id).Error
}
