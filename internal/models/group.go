package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Group представляет учебную группу
type Group struct {
	ID        uuid.UUID      `json:"id" gorm:"type:text;primaryKey"`
	Name      string         `json:"name" gorm:"type:text;not null"`
	Subject   string         `json:"subject" gorm:"type:text"`
	Grade     int            `json:"grade" gorm:"type:integer"`
	Level     int            `json:"level" gorm:"type:integer"`
	TeacherID uuid.UUID      `json:"teacher_id" gorm:"type:text;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Связи
	Teacher User          `json:"teacher" gorm:"foreignKey:TeacherID"`
	Members []GroupMember `json:"members" gorm:"foreignKey:GroupID"`
}

// GroupMember представляет участника группы
type GroupMember struct {
	ID        uuid.UUID      `json:"id" gorm:"type:text;primaryKey"`
	GroupID   uuid.UUID      `json:"group_id" gorm:"type:text;not null"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:text;not null"`
	Role      string         `json:"role" gorm:"type:text;default:'student'"` // student|assistant
	JoinedAt  time.Time      `json:"joined_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Связи
	Group Group `json:"group" gorm:"foreignKey:GroupID"`
	User  User  `json:"user" gorm:"foreignKey:UserID"`
}
