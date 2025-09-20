package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// Storage представляет файловое хранилище
type Storage struct {
	basePath       string
	maxFileSize    int64
	maxUserStorage int64
}

// NewStorage создает новое файловое хранилище
func NewStorage(basePath string, maxFileSize, maxUserStorage int64) (*Storage, error) {
	// Создаем базовую директорию
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Storage{
		basePath:       basePath,
		maxFileSize:    maxFileSize,
		maxUserStorage: maxUserStorage,
	}, nil
}

// SaveFile сохраняет загруженный файл
func (s *Storage) SaveFile(file *multipart.FileHeader, userID uuid.UUID, category string) (string, error) {
	// Проверяем размер файла
	if file.Size > s.maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size")
	}

	// Проверяем общий размер файлов пользователя
	if err := s.checkUserStorage(userID); err != nil {
		return "", err
	}

	// Генерируем уникальное имя файла
	fileExt := filepath.Ext(file.Filename)
	fileName := uuid.New().String() + fileExt

	// Создаем путь для файла
	filePath := filepath.Join(s.basePath, "users", userID.String(), category, fileName)

	// Создаем директории
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create file directory: %w", err)
	}

	// Открываем исходный файл
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Создаем целевой файл
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Создаем превью для изображений
	if strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		if err := s.createThumbnail(filePath); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			fmt.Printf("Failed to create thumbnail: %v\n", err)
		}
	}

	return filePath, nil
}

// createThumbnail создает миниатюру изображения
func (s *Storage) createThumbnail(filePath string) error {
	// Открываем изображение
	img, err := imaging.Open(filePath)
	if err != nil {
		return err
	}

	// Создаем миниатюру 300x300
	thumbnail := imaging.Resize(img, 300, 300, imaging.Lanczos)

	// Сохраняем миниатюру
	thumbPath := strings.Replace(filePath, filepath.Ext(filePath), "_thumb.jpg", 1)
	return imaging.Save(thumbnail, thumbPath, imaging.JPEGQuality(85))
}

// checkUserStorage проверяет общий размер файлов пользователя
func (s *Storage) checkUserStorage(userID uuid.UUID) error {
	userDir := filepath.Join(s.basePath, "users", userID.String())

	var totalSize int64
	err := filepath.Walk(userDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to calculate user storage: %w", err)
	}

	if totalSize > s.maxUserStorage {
		return fmt.Errorf("user storage limit exceeded")
	}

	return nil
}

// DeleteFile удаляет файл
func (s *Storage) DeleteFile(filePath string) error {
	// Удаляем основной файл
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Удаляем миниатюру если она существует
	thumbPath := strings.Replace(filePath, filepath.Ext(filePath), "_thumb.jpg", 1)
	if err := os.Remove(thumbPath); err != nil && !os.IsNotExist(err) {
		// Логируем ошибку, но не прерываем выполнение
		fmt.Printf("Failed to delete thumbnail: %v\n", err)
	}

	return nil
}

// GetFileInfo возвращает информацию о файле
func (s *Storage) GetFileInfo(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}

// GetThumbnailPath возвращает путь к миниатюре файла
func (s *Storage) GetThumbnailPath(filePath string) string {
	return strings.Replace(filePath, filepath.Ext(filePath), "_thumb.jpg", 1)
}

// CleanupOldFiles удаляет старые временные файлы
func (s *Storage) CleanupOldFiles() error {
	tempDir := filepath.Join(s.basePath, "temp")

	return filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Удаляем файлы старше 24 часов
		if !info.IsDir() && time.Since(info.ModTime()) > 24*time.Hour {
			return os.Remove(path)
		}

		return nil
	})
}
