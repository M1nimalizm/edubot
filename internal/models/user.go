package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole определяет роли пользователей
type UserRole string

const (
	RoleGuest     UserRole = "guest"
	RoleStudent   UserRole = "student"
	RoleTeacher   UserRole = "teacher"
)

// User представляет пользователя системы
type User struct {
	ID         uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	TelegramID int64     `json:"telegram_id" gorm:"uniqueIndex;not null"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Role       UserRole  `json:"role" gorm:"default:'guest'"`
	Phone      string    `json:"phone"`
	Grade      int       `json:"grade"` // 10 или 11 класс
	Subjects   string    `json:"subjects"` // JSON массив предметов
	InviteCode string    `json:"invite_code" gorm:"uniqueIndex"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// TrialRequest представляет заявку на пробное занятие
type TrialRequest struct {
	ID         uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Name       string    `json:"name" gorm:"not null"`
	Grade      int       `json:"grade" gorm:"not null"`
	Subject    string    `json:"subject" gorm:"not null"` // "physics", "math", "both"
	Level      int       `json:"level" gorm:"not null"` // 1-5
	Comment    string    `json:"comment"`
	Phone      string    `json:"phone"`
	TelegramID int64     `json:"telegram_id"`
	Status     string    `json:"status" gorm:"default:'pending'"` // "pending", "contacted", "converted", "rejected"
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Assignment представляет домашнее задание
type Assignment struct {
	ID          uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Subject     string    `json:"subject" gorm:"not null"` // "physics", "math"
	Deadline    time.Time `json:"deadline"`
	CreatedBy   uuid.UUID `json:"created_by" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Связи
	Creator     User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	Attachments []Attachment   `json:"attachments" gorm:"foreignKey:AssignmentID"`
	Submissions []Submission   `json:"submissions" gorm:"foreignKey:AssignmentID"`
	UserAssignments []UserAssignment `json:"user_assignments" gorm:"foreignKey:AssignmentID"`
}

// UserAssignment связывает пользователей с заданиями
type UserAssignment struct {
	ID           uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:text"`
	AssignmentID uuid.UUID `json:"assignment_id" gorm:"type:text"`
	AssignedAt   time.Time `json:"assigned_at"`
	
	// Связи
	User       User       `json:"user" gorm:"foreignKey:UserID"`
	Assignment Assignment `json:"assignment" gorm:"foreignKey:AssignmentID"`
}

// Submission представляет решение задания от ученика
type Submission struct {
	ID           uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	AssignmentID uuid.UUID `json:"assignment_id" gorm:"type:text"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:text"`
	Status       string    `json:"status" gorm:"default:'submitted'"` // "submitted", "graded", "returned"
	Grade        int       `json:"grade"` // 1-5
	Comments     string    `json:"comments"`
	SubmittedAt  time.Time `json:"submitted_at"`
	GradedAt     *time.Time `json:"graded_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Связи
	Assignment Assignment   `json:"assignment" gorm:"foreignKey:AssignmentID"`
	User       User         `json:"user" gorm:"foreignKey:UserID"`
	Files      []Attachment `json:"files" gorm:"foreignKey:SubmissionID"`
}

// Attachment представляет прикрепленный файл
type Attachment struct {
	ID           uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	FileName     string    `json:"file_name"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	AssignmentID *uuid.UUID `json:"assignment_id" gorm:"type:text"`
	SubmissionID *uuid.UUID `json:"submission_id" gorm:"type:text"`
	ContentID    *uuid.UUID `json:"content_id" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	
	// Связи
	Assignment *Assignment `json:"assignment,omitempty" gorm:"foreignKey:AssignmentID"`
	Submission *Submission `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
	Content    *Content    `json:"content,omitempty" gorm:"foreignKey:ContentID"`
}

// Content представляет образовательный контент
type Content struct {
	ID          uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Type        string    `json:"type" gorm:"not null"` // "video", "document", "image", "test"
	Category    string    `json:"category"` // "tips", "solutions", "tests", "demos", "reference"
	Tags        string    `json:"tags"` // JSON массив тегов
	CreatedBy   uuid.UUID `json:"created_by" gorm:"type:text"`
	IsPublic    bool      `json:"is_public" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Связи
	Creator     User         `json:"creator" gorm:"foreignKey:CreatedBy"`
	Attachments []Attachment `json:"attachments" gorm:"foreignKey:ContentID"`
}

// ContentView представляет просмотр контента пользователем
type ContentView struct {
	ID        uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	ContentID uuid.UUID `json:"content_id" gorm:"type:text"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:text"`
	ViewedAt  time.Time `json:"viewed_at"`
	
	// Связи
	Content Content `json:"content" gorm:"foreignKey:ContentID"`
	User    User    `json:"user" gorm:"foreignKey:UserID"`
}
