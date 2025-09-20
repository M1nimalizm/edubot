package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config содержит все настройки приложения
type Config struct {
	// Server
	Port string
	Host string

	// Database
	DBPath string

	// Telegram
	TelegramBotToken   string
	TelegramWebhookURL string
	TeacherTelegramID  int64

	// File Storage
	UploadPath     string
	MaxFileSize    int64
	MaxUserStorage int64

	// Security
	JWTSecret     string
	JWTExpiration time.Duration
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Загружаем .env файл если он существует
	_ = godotenv.Load()

	config := &Config{
		Port:               getEnv("PORT", "8080"),
		Host:               getEnv("HOST", "localhost"),
		DBPath:             getEnv("DB_PATH", "./data/edubot.db"),
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWebhookURL: getEnv("TELEGRAM_WEBHOOK_URL", ""),
		UploadPath:         getEnv("UPLOAD_PATH", "./uploads"),
		JWTSecret:          getEnv("JWT_SECRET", "default_secret_key"),
		JWTExpiration:      24 * time.Hour,
	}

	// Парсим числовые значения
	if maxFileSize, err := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "52428800"), 10, 64); err == nil {
		config.MaxFileSize = maxFileSize
	} else {
		config.MaxFileSize = 50 * 1024 * 1024 // 50MB по умолчанию
	}

	if maxUserStorage, err := strconv.ParseInt(getEnv("MAX_USER_STORAGE", "524288000"), 10, 64); err == nil {
		config.MaxUserStorage = maxUserStorage
	} else {
		config.MaxUserStorage = 500 * 1024 * 1024 // 500MB по умолчанию
	}

	if teacherID, err := strconv.ParseInt(getEnv("TEACHER_TELEGRAM_ID", "0"), 10, 64); err == nil {
		config.TeacherTelegramID = teacherID
	}

	return config, nil
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
