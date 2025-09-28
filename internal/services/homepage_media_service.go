package services

import (
	"edubot/internal/models"
	"edubot/internal/repository"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HomepageMediaService interface {
	UploadFile(file *multipart.FileHeader, mediaType models.HomepageMediaType) (*models.HomepageMedia, error)
	GetActiveMedia(mediaType models.HomepageMediaType) (*models.HomepageMedia, error)
	GetMediaByID(id uuid.UUID) (*models.HomepageMedia, error)
	ListMedia() ([]*models.HomepageMedia, error)
	DeleteMedia(id uuid.UUID) error
	SetActiveMedia(mediaType models.HomepageMediaType, mediaID uuid.UUID) error
	GetMediaURL(media *models.HomepageMedia) string
}

type homepageMediaService struct {
	mediaRepo repository.HomepageMediaRepository
	baseURL   string
	uploadDir string
}

func NewHomepageMediaService(mediaRepo repository.HomepageMediaRepository, baseURL, uploadDir string) HomepageMediaService {
	// Создаем директорию для загрузок, если её нет
	os.MkdirAll(uploadDir, 0755)
	
	return &homepageMediaService{
		mediaRepo: mediaRepo,
		baseURL:   baseURL,
		uploadDir: uploadDir,
	}
}

func (s *homepageMediaService) UploadFile(file *multipart.FileHeader, mediaType models.HomepageMediaType) (*models.HomepageMedia, error) {
	// Генерируем уникальное имя файла
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", string(mediaType), time.Now().Unix(), ext)
	
	// Создаем путь для сохранения
	filePath := filepath.Join(s.uploadDir, filename)
	
	// Открываем загружаемый файл
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()
	
	// Создаем файл назначения
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()
	
	// Копируем содержимое
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}
	
	// Получаем информацию о файле
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	
	// Определяем размеры изображения (если это изображение)
	var width, height int
	if strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		// Попробуем определить размеры изображения без внешних библиотек
		// Для простоты пока оставляем 0, 0 - размеры можно добавить позже
		width, height = 0, 0
	}
	
	// Создаем запись в базе данных
	media := &models.HomepageMedia{
		ID:        uuid.New(),
		Type:      mediaType,
		Filename:  filename,
		Path:      filePath,
		URL:       fmt.Sprintf("%s/media/homepage/%s", s.baseURL, filename),
		Size:      fileInfo.Size(),
		MimeType:  file.Header.Get("Content-Type"),
		Width:     width,
		Height:    height,
		IsActive:  false, // По умолчанию неактивен
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Сохраняем в базу данных
	if err := s.mediaRepo.Create(media); err != nil {
		// Удаляем файл, если не удалось сохранить в БД
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save media to database: %w", err)
	}
	
	return media, nil
}

func (s *homepageMediaService) GetActiveMedia(mediaType models.HomepageMediaType) (*models.HomepageMedia, error) {
	return s.mediaRepo.GetActiveByType(mediaType)
}

func (s *homepageMediaService) GetMediaByID(id uuid.UUID) (*models.HomepageMedia, error) {
	return s.mediaRepo.GetByID(id)
}

func (s *homepageMediaService) ListMedia() ([]*models.HomepageMedia, error) {
	return s.mediaRepo.List()
}

func (s *homepageMediaService) DeleteMedia(id uuid.UUID) error {
	// Получаем информацию о файле
	media, err := s.mediaRepo.GetByID(id)
	if err != nil {
		return err
	}
	
	// Удаляем файл с диска
	if err := os.Remove(media.Path); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		fmt.Printf("Warning: failed to delete file %s: %v\n", media.Path, err)
	}
	
	// Удаляем запись из базы данных
	return s.mediaRepo.Delete(id)
}

func (s *homepageMediaService) SetActiveMedia(mediaType models.HomepageMediaType, mediaID uuid.UUID) error {
	return s.mediaRepo.SetActive(mediaType, mediaID)
}

func (s *homepageMediaService) GetMediaURL(media *models.HomepageMedia) string {
	return media.URL
}
