package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Assignment - задание для ученика
type Assignment struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Subject     string    `json:"subject" gorm:"not null"` // physics, math, both
	Grade       int       `json:"grade" gorm:"not null"`   // 10, 11
	Level       int       `json:"level" gorm:"not null"`   // 1-5
	
	// Связи
	TeacherID   uuid.UUID `json:"teacher_id" gorm:"type:uuid;not null"`
	StudentID   uuid.UUID `json:"student_id" gorm:"type:uuid;not null"`
	
	// Дедлайны
	CreatedAt   time.Time `json:"created_at"`
	DueDate     time.Time `json:"due_date"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	
	// Статус
	Status      string    `json:"status" gorm:"default:'assigned'"` // assigned, completed, overdue
	
	// Связи
	Teacher     User      `json:"teacher" gorm:"foreignKey:TeacherID"`
	Student     User      `json:"student" gorm:"foreignKey:StudentID"`
	Comments    []Comment `json:"comments" gorm:"foreignKey:AssignmentID"`
	
	// Мягкое удаление
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Comment - комментарий к заданию
type Comment struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Content     string    `json:"content" gorm:"not null"`
	AuthorType  string    `json:"author_type" gorm:"not null"` // teacher, student
	
	// Связи
	AssignmentID uuid.UUID `json:"assignment_id" gorm:"type:uuid;not null"`
	AuthorID     uuid.UUID `json:"author_id" gorm:"type:uuid;not null"`
	
	CreatedAt   time.Time `json:"created_at"`
	
	// Связи
	Assignment  Assignment `json:"assignment" gorm:"foreignKey:AssignmentID"`
	Author      User       `json:"author" gorm:"foreignKey:AuthorID"`
	
	// Мягкое удаление
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Content - учебный материал
type Content struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Type        string    `json:"type" gorm:"not null"` // video, pdf, link, text
	URL         string    `json:"url"`
	Content     string    `json:"content"` // для текстового контента
	
	Subject     string    `json:"subject" gorm:"not null"` // physics, math, both
	Grade       int       `json:"grade" gorm:"not null"`   // 10, 11
	Level       int       `json:"level" gorm:"not null"`   // 1-5
	
	// Связи
	TeacherID   uuid.UUID `json:"teacher_id" gorm:"type:uuid;not null"`
	
	CreatedAt   time.Time `json:"created_at"`
	
	// Связи
	Teacher     User      `json:"teacher" gorm:"foreignKey:TeacherID"`
	
	// Мягкое удаление
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// StudentProgress - прогресс ученика
type StudentProgress struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	
	// Связи
	StudentID   uuid.UUID `json:"student_id" gorm:"type:uuid;not null"`
	Subject     string    `json:"subject" gorm:"not null"` // physics, math
	
	// Прогресс
	Level       int       `json:"level" gorm:"not null"` // 1-5
	CompletedAssignments int `json:"completed_assignments" gorm:"default:0"`
	TotalAssignments    int `json:"total_assignments" gorm:"default:0"`
	AverageScore        float64 `json:"average_score" gorm:"default:0"`
	
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Связи
	Student     User      `json:"student" gorm:"foreignKey:StudentID"`
	
	// Мягкое удаление
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
