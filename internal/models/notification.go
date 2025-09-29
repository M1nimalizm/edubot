package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType определяет типы уведомлений
type NotificationType string

const (
	NotificationTypeNewAssignment    NotificationType = "new_assignment"
	NotificationTypeDeadlineReminder NotificationType = "deadline_reminder"
	NotificationTypeOverdue          NotificationType = "overdue"
	NotificationTypeGradeReceived    NotificationType = "grade_received"
	NotificationTypeNewMessage       NotificationType = "new_message"
	NotificationTypeGroupInvite      NotificationType = "group_invite"
)

// NotificationChannel определяет каналы доставки
type NotificationChannel string

const (
	NotificationChannelBot   NotificationChannel = "bot"
	NotificationChannelInApp NotificationChannel = "inapp"
	NotificationChannelEmail NotificationChannel = "email"
)

// NotificationStatus определяет статусы уведомлений
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusRead    NotificationStatus = "read"
)

// Notification представляет уведомление пользователю
type Notification struct {
	ID        uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID           `json:"user_id" gorm:"type:uuid;not null"`
	Type      NotificationType    `json:"type" gorm:"type:varchar(30);not null"`
	Title     string              `json:"title" gorm:"not null"`
	Message   string              `json:"message" gorm:"not null"`
	Payload   string              `json:"payload" gorm:"type:text"` // JSON с дополнительными данными
	Channel   NotificationChannel `json:"channel" gorm:"type:varchar(10);not null"`
	Status    NotificationStatus  `json:"status" gorm:"type:varchar(10);default:'pending'"`
	SentAt    *time.Time          `json:"sent_at,omitempty"`
	ReadAt    *time.Time          `json:"read_at,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `json:"deleted_at,omitempty" gorm:"index"`

	// Связи
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Draft представляет черновик отправки ДЗ
type Draft struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	AssignmentTargetID *uuid.UUID     `json:"assignment_target_id,omitempty" gorm:"type:uuid"`
	StudentID          uuid.UUID      `json:"student_id" gorm:"type:uuid;not null"`
	Text               *string        `json:"text,omitempty"`
	MediaIDs           string         `json:"media_ids" gorm:"type:text"` // JSON массив ID медиа
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Связи
	AssignmentTarget *AssignmentTarget `json:"assignment_target,omitempty" gorm:"foreignKey:AssignmentTargetID"`
	Student          User              `json:"student" gorm:"foreignKey:StudentID"`
}
