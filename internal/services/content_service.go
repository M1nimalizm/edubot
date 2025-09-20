package services

import (
	"fmt"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/storage"
	"mime/multipart"

	"github.com/google/uuid"
)

// ContentService представляет сервис для работы с образовательным контентом
type ContentService struct {
	contentRepo    *repository.ContentRepository
	attachmentRepo *repository.AttachmentRepository
	storage        *storage.Storage
}

// NewContentService создает новый сервис контента
func NewContentService(
	contentRepo *repository.ContentRepository,
	attachmentRepo *repository.AttachmentRepository,
	storage *storage.Storage,
) *ContentService {
	return &ContentService{
		contentRepo:    contentRepo,
		attachmentRepo: attachmentRepo,
		storage:        storage,
	}
}

// CreateContentRequest представляет запрос на создание контента
type CreateContentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Tags        string `json:"tags"`
	IsPublic    bool   `json:"is_public"`
}

// CreateContent создает новый контент
func (s *ContentService) CreateContent(req *CreateContentRequest, creatorID uuid.UUID, files []*multipart.FileHeader) (*models.Content, error) {
	// Создаем контент
	content := &models.Content{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Category:    req.Category,
		Tags:        req.Tags,
		CreatedBy:   creatorID,
		IsPublic:    req.IsPublic,
	}

	if err := s.contentRepo.Create(content); err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	// Сохраняем файлы
	for _, file := range files {
		filePath, err := s.storage.SaveFile(file, creatorID, "content")
		if err != nil {
			return nil, fmt.Errorf("failed to save file %s: %w", file.Filename, err)
		}

		attachment := &models.Attachment{
			FileName:     file.Filename,
			OriginalName: file.Filename,
			FilePath:     filePath,
			FileSize:     file.Size,
			MimeType:     file.Header.Get("Content-Type"),
			ContentID:    &content.ID,
		}

		if err := s.attachmentRepo.Create(attachment); err != nil {
			return nil, fmt.Errorf("failed to create attachment: %w", err)
		}
	}

	return content, nil
}

// GetContent получает контент по ID
func (s *ContentService) GetContent(contentID uuid.UUID) (*models.Content, error) {
	return s.contentRepo.GetByID(contentID)
}

// ListContent получает список всего публичного контента
func (s *ContentService) ListContent() ([]models.Content, error) {
	return s.contentRepo.List()
}

// ListContentByType получает контент по типу
func (s *ContentService) ListContentByType(contentType string) ([]models.Content, error) {
	return s.contentRepo.ListByType(contentType)
}

// ListContentByCategory получает контент по категории
func (s *ContentService) ListContentByCategory(category string) ([]models.Content, error) {
	return s.contentRepo.ListByCategory(category)
}

// SearchContent выполняет поиск контента
func (s *ContentService) SearchContent(query string) ([]models.Content, error) {
	return s.contentRepo.Search(query)
}

// MarkAsViewed отмечает контент как просмотренный
func (s *ContentService) MarkAsViewed(contentID, userID uuid.UUID) error {
	return s.contentRepo.MarkAsViewed(contentID, userID)
}

// GetViewedContent получает просмотренный пользователем контент
func (s *ContentService) GetViewedContent(userID uuid.UUID) ([]models.Content, error) {
	return s.contentRepo.GetViewedContent(userID)
}

// UpdateContent обновляет контент
func (s *ContentService) UpdateContent(contentID uuid.UUID, req *CreateContentRequest) error {
	content, err := s.contentRepo.GetByID(contentID)
	if err != nil {
		return fmt.Errorf("content not found: %w", err)
	}

	content.Title = req.Title
	content.Description = req.Description
	content.Type = req.Type
	content.Category = req.Category
	content.Tags = req.Tags
	content.IsPublic = req.IsPublic

	return s.contentRepo.Update(content)
}

// DeleteContent удаляет контент
func (s *ContentService) DeleteContent(contentID uuid.UUID) error {
	// Получаем контент для удаления файлов
	content, err := s.contentRepo.GetByID(contentID)
	if err != nil {
		return fmt.Errorf("content not found: %w", err)
	}

	// Удаляем файлы
	for _, attachment := range content.Attachments {
		if err := s.storage.DeleteFile(attachment.FilePath); err != nil {
			fmt.Printf("Failed to delete file %s: %v\n", attachment.FilePath, err)
		}
	}

	return s.contentRepo.Delete(contentID)
}

// GetContentCategories возвращает список доступных категорий
func (s *ContentService) GetContentCategories() []string {
	return []string{
		"tips",      // Видео-лайфхаки
		"solutions", // Разборы заданий
		"tests",     // Интерактивные тесты
		"demos",     // Демонстрации с уроков
		"reference", // Справочные материалы
	}
}

// GetContentTypes возвращает список доступных типов контента
func (s *ContentService) GetContentTypes() []string {
	return []string{
		"video",    // Видео
		"document", // Документ
		"image",    // Изображение
		"test",     // Тест
	}
}

// GetContentStats возвращает статистику контента
func (s *ContentService) GetContentStats() (map[string]interface{}, error) {
	// Получаем общее количество контента
	allContent, err := s.contentRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get content list: %w", err)
	}

	stats := map[string]interface{}{
		"total_content": len(allContent),
		"by_type":       make(map[string]int),
		"by_category":   make(map[string]int),
	}

	// Подсчитываем по типам и категориям
	for _, content := range allContent {
		// По типам
		if count, ok := stats["by_type"].(map[string]int)[content.Type]; ok {
			stats["by_type"].(map[string]int)[content.Type] = count + 1
		} else {
			stats["by_type"].(map[string]int)[content.Type] = 1
		}

		// По категориям
		if count, ok := stats["by_category"].(map[string]int)[content.Category]; ok {
			stats["by_category"].(map[string]int)[content.Category] = count + 1
		} else {
			stats["by_category"].(map[string]int)[content.Category] = 1
		}
	}

	return stats, nil
}
