package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatThreadType определяет типы чатов
type ChatThreadType string

const (
	ChatThreadTypeStudentTeacher ChatThreadType = "student_teacher"
	ChatThreadTypeGroup          ChatThreadType = "group"
)

// ChatThread представляет чат между учеником и учителем или групповой чат
type ChatThread struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Type          ChatThreadType `json:"type" gorm:"type:varchar(20);not null"`
	StudentID     *uuid.UUID     `json:"student_id,omitempty" gorm:"type:uuid"` // Для личного чата
	GroupID       *uuid.UUID     `json:"group_id,omitempty" gorm:"type:uuid"`   // Для группового чата
	TeacherID     uuid.UUID      `json:"teacher_id" gorm:"type:uuid;not null"`
	LastMessageAt *time.Time     `json:"last_message_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Связи
	Student  User      `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	Group    Group     `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	Teacher  User      `json:"teacher" gorm:"foreignKey:TeacherID"`
	Messages []Message `json:"messages" gorm:"foreignKey:ThreadID"`
}

// MessageKind определяет типы сообщений
type MessageKind string

const (
	MessageKindMessage MessageKind = "message"
	MessageKindSystem  MessageKind = "system"
	MessageKindGrade   MessageKind = "grade"
)

// Message представляет сообщение в чате
type Message struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ThreadID  uuid.UUID      `json:"thread_id" gorm:"type:uuid;not null"`
	AuthorID  uuid.UUID      `json:"author_id" gorm:"type:uuid;not null"`
	Text      *string        `json:"text,omitempty"`
	Kind      MessageKind    `json:"kind" gorm:"type:varchar(20);default:'message'"`
	CreatedAt time.Time      `json:"created_at"`
	EditedAt  *time.Time     `json:"edited_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Связи
	Thread ChatThread `json:"thread" gorm:"foreignKey:ThreadID"`
	Author User       `json:"author" gorm:"foreignKey:AuthorID"`
	Media  []Media    `json:"media" gorm:"many2many:message_media;"`
}
