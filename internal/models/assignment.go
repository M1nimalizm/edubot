package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment - комментарий к существующему Assignment
type Comment struct {
	ID         uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Content    string    `json:"content" gorm:"not null"`
	AuthorType string    `json:"author_type" gorm:"not null"` // teacher, student

    AssignmentID uuid.UUID `json:"assignment_id" gorm:"type:text;not null"`
	AuthorID uuid.UUID `json:"author_id" gorm:"type:text;not null"`

	CreatedAt time.Time `json:"created_at"`

    Assignment Assignment `json:"assignment" gorm:"foreignKey:AssignmentID"`
	Author User `json:"author" gorm:"foreignKey:AuthorID"`

	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// StudentProgress - прогресс ученика
type StudentProgress struct {
	ID                    uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	StudentID             uuid.UUID `json:"student_id" gorm:"type:text;not null"`
	Subject               string    `json:"subject" gorm:"not null"`
	Level                 int       `json:"level" gorm:"not null"`
	CompletedAssignments  int       `json:"completed_assignments" gorm:"default:0"`
	TotalAssignments      int       `json:"total_assignments" gorm:"default:0"`
	AverageScore          float64   `json:"average_score" gorm:"default:0"`
	LastActivity          time.Time `json:"last_activity"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	Student               User      `json:"student" gorm:"foreignKey:StudentID"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
