package database

import (
	"fmt"
	"os"
	"path/filepath"

	"edubot/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
    "github.com/google/uuid"
)

// Database представляет подключение к базе данных
type Database struct {
	DB *gorm.DB
}

// NewDatabase создает новое подключение к базе данных
func NewDatabase(dbPath string) (*Database, error) {
	// Создаем директорию для базы данных если она не существует
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Подключаемся к SQLite базе данных
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	database := &Database{DB: db}

	// Автомиграция моделей
	if err := database.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return database, nil
}

// Migrate выполняет миграцию базы данных
func (d *Database) Migrate() error {
	return d.DB.AutoMigrate(
		&models.User{},
		&models.TrialRequest{},
		&models.Assignment{},
		&models.Comment{},
		&models.Content{},
		&models.StudentProgress{},
		&models.Media{},
		&models.MediaAccess{},
		&models.MediaView{},
	)
}

// Close закрывает подключение к базе данных
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateDefaultTeacher создает пользователя-преподавателя по умолчанию
func (d *Database) CreateDefaultTeacher(telegramID int64) error {
	var user models.User
	result := d.DB.Where("telegram_id = ?", telegramID).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// Создаем преподавателя
        teacher := models.User{
            ID:         uuid.New(),
			TelegramID: telegramID,
			FirstName:  "Александр",
			LastName:   "Пугачев",
			Role:       models.RoleTeacher,
			Username:   "pugachev_teacher",
		}

        if err := d.DB.Create(&teacher).Error; err != nil {
			return fmt.Errorf("failed to create default teacher: %w", err)
		}
	}

	return nil
}
