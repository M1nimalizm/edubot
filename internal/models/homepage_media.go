package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HomepageMediaType тип медиафайла для главной страницы
type HomepageMediaType string

const (
	HomepageMediaTypeTeacherPhoto HomepageMediaType = "teacher_photo"
	HomepageMediaTypeWelcomeVideo HomepageMediaType = "welcome_video"
	HomepageMediaTypeHeroImage    HomepageMediaType = "hero_image"
)

// HomepageMedia медиафайлы для главной страницы
type HomepageMedia struct {
	ID        uuid.UUID         `json:"id" gorm:"type:text;primaryKey"`
	Type      HomepageMediaType `json:"type" gorm:"type:text;not null"`
	Filename  string            `json:"filename" gorm:"type:text;not null"`
	Path      string            `json:"path" gorm:"type:text;not null"`
	URL       string            `json:"url" gorm:"type:text;not null"`
	Size      int64             `json:"size"`
	MimeType  string            `json:"mime_type" gorm:"type:text"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	IsActive  bool              `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt gorm.DeletedAt    `json:"-" gorm:"index"`
}
