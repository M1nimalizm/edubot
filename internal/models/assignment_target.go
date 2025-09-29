package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssignmentTargetStatus определяет статусы индивидуального задания
type AssignmentTargetStatus string

const (
	AssignmentTargetStatusPending   AssignmentTargetStatus = "pending"
	AssignmentTargetStatusSubmitted AssignmentTargetStatus = "submitted"
	AssignmentTargetStatusGraded    AssignmentTargetStatus = "graded"
	AssignmentTargetStatusOverdue   AssignmentTargetStatus = "overdue"
)

// AssignmentTarget представляет индивидуальное задание для конкретного ученика
// Создается при раздаче группового Assignment
type AssignmentTarget struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey"`
	AssignmentID uuid.UUID              `json:"assignment_id" gorm:"type:uuid;not null"`
	StudentID    uuid.UUID              `json:"student_id" gorm:"type:uuid;not null"`
	Status       AssignmentTargetStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Score        *float64               `json:"score,omitempty"` // Оценка от учителя
	SubmittedAt  *time.Time             `json:"submitted_at,omitempty"`
	GradedAt     *time.Time             `json:"graded_at,omitempty"`
	IsLate       bool                   `json:"is_late" gorm:"default:false"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeletedAt    gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`

	// Связи
	Assignment  Assignment   `json:"assignment" gorm:"foreignKey:AssignmentID"`
	Student     User         `json:"student" gorm:"foreignKey:StudentID"`
	Submissions []Submission `json:"submissions" gorm:"foreignKey:AssignmentTargetID"`
	Feedbacks   []Feedback   `json:"feedbacks" gorm:"foreignKey:AssignmentTargetID"`
}

// Feedback представляет оценку и комментарий учителя к заданию
type Feedback struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	AssignmentTargetID uuid.UUID      `json:"assignment_target_id" gorm:"type:uuid;not null"`
	TeacherID          uuid.UUID      `json:"teacher_id" gorm:"type:uuid;not null"`
	Text               string         `json:"text"`
	Score              *float64       `json:"score,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Связи
	AssignmentTarget AssignmentTarget `json:"assignment_target" gorm:"foreignKey:AssignmentTargetID"`
	Teacher          User             `json:"teacher" gorm:"foreignKey:TeacherID"`
	Media            []Media          `json:"media" gorm:"many2many:feedback_media;"`
}
