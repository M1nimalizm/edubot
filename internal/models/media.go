package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaType определяет типы медиафайлов
type MediaType string

const (
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeImage    MediaType = "image"
)

// MediaScope определяет область видимости медиа
type MediaScope string

const (
	MediaScopePublic  MediaScope = "public"
	MediaScopeStudent MediaScope = "student"
	MediaScopeTeacher MediaScope = "teacher"
	MediaScopePrivate MediaScope = "private"
)

// EntityType определяет тип сущности, к которой привязано медиа
type EntityType string

const (
	EntityTypeWelcomeVideo EntityType = "welcome_video"
	EntityTypeMaterial     EntityType = "material"
	EntityTypeAssignment   EntityType = "assignment"
	EntityTypeSubmission   EntityType = "submission"
	EntityTypeReview       EntityType = "review"
	EntityTypeContent      EntityType = "content"
)

// Media представляет медиафайл, хранящийся в Telegram
type Media struct {
	ID               uuid.UUID      `json:"id" gorm:"type:text;primaryKey"`
	TelegramFileID   string         `json:"telegram_file_id" gorm:"type:text;not null"`
	TelegramUniqueID string        `json:"telegram_unique_id" gorm:"type:text"`
	ChatID           int64          `json:"chat_id" gorm:"type:integer"`
	MessageID        int            `json:"message_id" gorm:"type:integer"`
	Type             MediaType      `json:"type" gorm:"type:text;not null"`
	MimeType         string         `json:"mime_type" gorm:"type:text"`
	Size             int64          `json:"size" gorm:"type:integer"`
	Caption          string         `json:"caption" gorm:"type:text"`
	OwnerID          uuid.UUID      `json:"owner_id" gorm:"type:text;not null"`
	Scope            MediaScope     `json:"scope" gorm:"type:text;default:'private'"`
	EntityType       EntityType     `json:"entity_type" gorm:"type:text"`
	EntityID         *uuid.UUID     `json:"entity_id" gorm:"type:text"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Связи
	Owner User `json:"owner" gorm:"foreignKey:OwnerID"`
}

// MediaAccess представляет права доступа к медиафайлу
type MediaAccess struct {
	ID        uuid.UUID `json:"id" gorm:"type:text;primaryKey"`
	MediaID   uuid.UUID `json:"media_id" gorm:"type:text;not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:text;not null"`
	Permission string    `json:"permission" gorm:"type:text;not null"` // "read", "write", "admin"
	CreatedAt time.Time `json:"created_at"`

	// Связи
	Media Media `json:"media" gorm:"foreignKey:MediaID"`
	User  User  `json:"user" gorm:"foreignKey:UserID"`
}

// MediaView представляет просмотр медиафайла пользователем
type MediaView struct {
	ID        uuid.UUID `json:"id" gorm:"type:text;primaryKey"`
	MediaID   uuid.UUID `json:"media_id" gorm:"type:text;not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:text;not null"`
	ViewedAt  time.Time `json:"viewed_at"`
	Duration  int       `json:"duration"` // Длительность просмотра в секундах

	// Связи
	Media Media `json:"media" gorm:"foreignKey:MediaID"`
	User  User  `json:"user" gorm:"foreignKey:UserID"`
}

// IsVideo проверяет, является ли медиафайл видео
func (m *Media) IsVideo() bool {
	return m.Type == MediaTypeVideo
}

// IsAudio проверяет, является ли медиафайл аудио
func (m *Media) IsAudio() bool {
	return m.Type == MediaTypeAudio
}

// IsDocument проверяет, является ли медиафайл документом
func (m *Media) IsDocument() bool {
	return m.Type == MediaTypeDocument
}

// IsImage проверяет, является ли медиафайл изображением
func (m *Media) IsImage() bool {
	return m.Type == MediaTypeImage
}

// IsPublic проверяет, является ли медиафайл публичным
func (m *Media) IsPublic() bool {
	return m.Scope == MediaScopePublic
}

// CanUserAccess проверяет, может ли пользователь получить доступ к медиафайлу
func (m *Media) CanUserAccess(userID uuid.UUID, userRole UserRole) bool {
	// Публичные медиа доступны всем
	if m.IsPublic() {
		return true
	}

	// Владелец всегда имеет доступ
	if m.OwnerID == userID {
		return true
	}

	// Учителя имеют доступ к медиа учеников
	if userRole == RoleTeacher && m.Scope == MediaScopeStudent {
		return true
	}

	// Ученики имеют доступ к своим медиа и публичным медиа учителей
	if userRole == RoleStudent && (m.Scope == MediaScopeStudent || m.Scope == MediaScopePublic) {
		return true
	}

	return false
}
