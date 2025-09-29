package services

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

// MediaService интерфейс для бизнес-логики медиафайлов
type MediaService interface {
	CreateMediaFromTelegram(fileID, uniqueID string, chatID int64, messageID int, mediaType models.MediaType, mimeType string, size int64, caption string, ownerID uuid.UUID, scope models.MediaScope, entityType models.EntityType, entityID *uuid.UUID) (*models.Media, error)
	GetMediaByID(id uuid.UUID) (*models.Media, error)
	GetMediaStream(id uuid.UUID, userID uuid.UUID) (io.ReadCloser, error)
	GetMediaThumbnail(id uuid.UUID, userID uuid.UUID) (io.ReadCloser, error)
	GetUserMedia(userID uuid.UUID) ([]*models.Media, error)
	GetPublicMedia() ([]*models.Media, error)
	GetEntityMedia(entityType models.EntityType, entityID uuid.UUID) ([]*models.Media, error)
	UpdateMedia(media *models.Media) error
	DeleteMedia(id uuid.UUID, userID uuid.UUID) error
	RecordView(mediaID, userID uuid.UUID, duration int) error
	GetMediaViews(mediaID uuid.UUID, limit int) ([]*models.MediaView, error)
	GrantMediaAccess(mediaID, userID uuid.UUID, permission string) error
	RevokeMediaAccess(mediaID, userID uuid.UUID) error
	CheckMediaAccess(mediaID, userID uuid.UUID) (bool, error)
}

type mediaService struct {
	mediaRepo      repository.MediaRepository
	userRepo       repository.UserRepository
	bot            *telegram.Bot
	assignmentRepo repository.AssignmentRepository
}

// NewMediaService создает новый сервис медиафайлов
func NewMediaService(mediaRepo repository.MediaRepository, userRepo repository.UserRepository, bot *telegram.Bot, assignmentRepo repository.AssignmentRepository) MediaService {
	return &mediaService{
		mediaRepo:      mediaRepo,
		userRepo:       userRepo,
		bot:            bot,
		assignmentRepo: assignmentRepo,
	}
}

// CreateMediaFromTelegram создает медиафайл из данных Telegram
func (s *mediaService) CreateMediaFromTelegram(fileID, uniqueID string, chatID int64, messageID int, mediaType models.MediaType, mimeType string, size int64, caption string, ownerID uuid.UUID, scope models.MediaScope, entityType models.EntityType, entityID *uuid.UUID) (*models.Media, error) {
	media := &models.Media{
		ID:               uuid.New(),
		TelegramFileID:   fileID,
		TelegramUniqueID: uniqueID,
		ChatID:           chatID,
		MessageID:        messageID,
		Type:             mediaType,
		MimeType:         mimeType,
		Size:             size,
		Caption:          caption,
		OwnerID:          ownerID,
		Scope:            scope,
		EntityType:       entityType,
		EntityID:         entityID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := s.mediaRepo.Create(media)
	if err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	return media, nil
}

// GetMediaByID получает медиафайл по ID
func (s *mediaService) GetMediaByID(id uuid.UUID) (*models.Media, error) {
	return s.mediaRepo.GetByID(id)
}

// GetMediaStream получает поток медиафайла для стриминга
func (s *mediaService) GetMediaStream(id uuid.UUID, userID uuid.UUID) (io.ReadCloser, error) {
	media, err := s.mediaRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("media not found: %w", err)
	}

	// user не нужен здесь, доступ проверяем через CheckMediaAccess

	// Проверяем права доступа (расширенная матрица)
	allowed, err := s.CheckMediaAccess(id, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("access denied")
	}

	// Получаем file_path от Telegram Bot API
	filePath, err := s.bot.GetFilePath(media.TelegramFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file path: %w", err)
	}

	// Скачиваем файл из Telegram
	resp, err := http.Get(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// GetMediaThumbnail получает миниатюру медиафайла
func (s *mediaService) GetMediaThumbnail(id uuid.UUID, userID uuid.UUID) (io.ReadCloser, error) {
	media, err := s.mediaRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("media not found: %w", err)
	}

	// user не нужен здесь, доступ проверяем через CheckMediaAccess

	// Проверяем права доступа (расширенная матрица)
	allowed, err := s.CheckMediaAccess(id, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("access denied")
	}

	// Для изображений возвращаем сам файл
	if media.IsImage() {
		return s.GetMediaStream(id, userID)
	}

	// Для видео пытаемся получить миниатюру
	if media.IsVideo() {
		// Здесь можно реализовать получение миниатюры через Bot API
		// Пока возвращаем основной поток
		return s.GetMediaStream(id, userID)
	}

	return nil, fmt.Errorf("thumbnail not available for this media type")
}

// GetUserMedia получает медиафайлы пользователя
func (s *mediaService) GetUserMedia(userID uuid.UUID) ([]*models.Media, error) {
	return s.mediaRepo.GetByOwnerID(userID)
}

// GetPublicMedia получает публичные медиафайлы
func (s *mediaService) GetPublicMedia() ([]*models.Media, error) {
	return s.mediaRepo.GetPublicMedia()
}

// GetEntityMedia получает медиафайлы сущности
func (s *mediaService) GetEntityMedia(entityType models.EntityType, entityID uuid.UUID) ([]*models.Media, error) {
	return s.mediaRepo.GetByEntity(entityType, entityID)
}

// UpdateMedia обновляет медиафайл
func (s *mediaService) UpdateMedia(media *models.Media) error {
	media.UpdatedAt = time.Now()
	return s.mediaRepo.Update(media)
}

// DeleteMedia удаляет медиафайл
func (s *mediaService) DeleteMedia(id uuid.UUID, userID uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Проверяем, что пользователь является владельцем
	if media.OwnerID != userID {
		return fmt.Errorf("access denied: not owner")
	}

	return s.mediaRepo.Delete(id)
}

// RecordView записывает просмотр медиафайла
func (s *mediaService) RecordView(mediaID, userID uuid.UUID, duration int) error {
	view := &models.MediaView{
		ID:       uuid.New(),
		MediaID:  mediaID,
		UserID:   userID,
		ViewedAt: time.Now(),
		Duration: duration,
	}

	return s.mediaRepo.AddView(view)
}

// GetMediaViews получает просмотры медиафайла
func (s *mediaService) GetMediaViews(mediaID uuid.UUID, limit int) ([]*models.MediaView, error) {
	return s.mediaRepo.GetRecentViews(mediaID, limit)
}

// GrantMediaAccess предоставляет доступ к медиафайлу
func (s *mediaService) GrantMediaAccess(mediaID, userID uuid.UUID, permission string) error {
	access := &models.MediaAccess{
		ID:         uuid.New(),
		MediaID:    mediaID,
		UserID:     userID,
		Permission: permission,
		CreatedAt:  time.Now(),
	}

	return s.mediaRepo.GrantAccess(access)
}

// RevokeMediaAccess отзывает доступ к медиафайлу
func (s *mediaService) RevokeMediaAccess(mediaID, userID uuid.UUID) error {
	return s.mediaRepo.RevokeAccess(mediaID, userID)
}

// CheckMediaAccess проверяет доступ к медиафайлу
func (s *mediaService) CheckMediaAccess(mediaID, userID uuid.UUID) (bool, error) {
	media, err := s.mediaRepo.GetByID(mediaID)
	if err != nil {
		return false, fmt.Errorf("media not found: %w", err)
	}

	// Public — всем
	if media.IsPublic() {
		return true, nil
	}

	// Владелец — всегда
	if media.OwnerID == userID {
		return true, nil
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	// Учительские материалы (teacher scope)
	if media.Scope == models.MediaScopeTeacher {
		if user.Role != models.RoleTeacher {
			return false, nil
		}
		// Если привязано к заданию/сабмишену/ревью — разрешаем только учителю этого задания
		if media.EntityType == models.EntityTypeAssignment && media.EntityID != nil {
			a, err := s.assignmentRepo.GetByID(*media.EntityID)
			if err != nil {
				return false, err
			}
			return a.TeacherID == userID, nil
		}
		if (media.EntityType == models.EntityTypeSubmission || media.EntityType == models.EntityTypeReview) && media.EntityID != nil {
			sub, err := s.assignmentRepo.GetSubmissionByID(*media.EntityID)
			if err != nil {
				return false, err
			}
			a, err := s.assignmentRepo.GetByID(sub.AssignmentID)
			if err != nil {
				return false, err
			}
			return a.TeacherID == userID, nil
		}
		// Иначе — любой учитель
		return true, nil
	}

	// Студенческие материалы (student scope)
	if media.Scope == models.MediaScopeStudent {
		// Assignment: студент задания или его учитель
		if media.EntityType == models.EntityTypeAssignment && media.EntityID != nil {
			a, err := s.assignmentRepo.GetByID(*media.EntityID)
			if err != nil {
				return false, err
			}
			return (a.StudentID != nil && *a.StudentID == userID) || a.TeacherID == userID, nil
		}
		// Submission/Review: автор сабмишена или учитель задания
		if (media.EntityType == models.EntityTypeSubmission || media.EntityType == models.EntityTypeReview) && media.EntityID != nil {
			sub, err := s.assignmentRepo.GetSubmissionByID(*media.EntityID)
			if err != nil {
				return false, err
			}
			a, err := s.assignmentRepo.GetByID(sub.AssignmentID)
			if err != nil {
				return false, err
			}
			return sub.UserID == userID || a.TeacherID == userID, nil
		}
		// Без сущности — по роли (ученик/учитель) запрещено, кроме владельца (уже проверили)
		return user.Role == models.RoleTeacher, nil
	}

	// Приватные: доступ только по явному доступу в ACL
	if media.Scope == models.MediaScopePrivate {
		accessList, err := s.mediaRepo.GetAccessList(media.ID)
		if err != nil {
			return false, err
		}
		for _, a := range accessList {
			if a.UserID == userID {
				return true, nil
			}
		}
		return false, nil
	}

	return false, nil
}
