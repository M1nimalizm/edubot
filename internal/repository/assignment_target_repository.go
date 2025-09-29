package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

type AssignmentTargetRepository interface {
	Create(target *models.AssignmentTarget) error
	GetByID(id uuid.UUID) (*models.AssignmentTarget, error)
	GetByAssignmentAndStudent(assignmentID, studentID uuid.UUID) (*models.AssignmentTarget, error)
	ListByStudent(studentID uuid.UUID) ([]*models.AssignmentTarget, error)
	ListByAssignment(assignmentID uuid.UUID) ([]*models.AssignmentTarget, error)
	ListByStatus(status models.AssignmentTargetStatus) ([]*models.AssignmentTarget, error)
	Update(target *models.AssignmentTarget) error
	Delete(id uuid.UUID) error

	// Методы для массовых операций
	CreateForGroup(assignmentID uuid.UUID, studentIDs []uuid.UUID) error
	MarkAsOverdue() error // Помечает просроченные задания
}

type assignmentTargetRepository struct {
	db *gorm.DB
}

func NewAssignmentTargetRepository(db *gorm.DB) AssignmentTargetRepository {
	return &assignmentTargetRepository{db: db}
}

func (r *assignmentTargetRepository) Create(target *models.AssignmentTarget) error {
	if target.ID == uuid.Nil {
		target.ID = uuid.New()
	}
	return r.db.Create(target).Error
}

func (r *assignmentTargetRepository) GetByID(id uuid.UUID) (*models.AssignmentTarget, error) {
	var target models.AssignmentTarget
	err := r.db.Preload("Assignment").Preload("Student").First(&target, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *assignmentTargetRepository) GetByAssignmentAndStudent(assignmentID, studentID uuid.UUID) (*models.AssignmentTarget, error) {
	var target models.AssignmentTarget
	err := r.db.Preload("Assignment").Preload("Student").
		Where("assignment_id = ? AND student_id = ?", assignmentID, studentID).
		First(&target).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *assignmentTargetRepository) ListByStudent(studentID uuid.UUID) ([]*models.AssignmentTarget, error) {
	var targets []*models.AssignmentTarget
	err := r.db.Preload("Assignment").Preload("Assignment.Teacher").
		Where("student_id = ?", studentID).
		Order("created_at DESC").
		Find(&targets).Error
	return targets, err
}

func (r *assignmentTargetRepository) ListByAssignment(assignmentID uuid.UUID) ([]*models.AssignmentTarget, error) {
	var targets []*models.AssignmentTarget
	err := r.db.Preload("Student").
		Where("assignment_id = ?", assignmentID).
		Order("created_at DESC").
		Find(&targets).Error
	return targets, err
}

func (r *assignmentTargetRepository) ListByStatus(status models.AssignmentTargetStatus) ([]*models.AssignmentTarget, error) {
	var targets []*models.AssignmentTarget
	err := r.db.Preload("Assignment").Preload("Student").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&targets).Error
	return targets, err
}

func (r *assignmentTargetRepository) Update(target *models.AssignmentTarget) error {
	return r.db.Save(target).Error
}

func (r *assignmentTargetRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.AssignmentTarget{}, "id = ?", id).Error
}

func (r *assignmentTargetRepository) CreateForGroup(assignmentID uuid.UUID, studentIDs []uuid.UUID) error {
	var targets []models.AssignmentTarget
	for _, studentID := range studentIDs {
		targets = append(targets, models.AssignmentTarget{
			ID:           uuid.New(),
			AssignmentID: assignmentID,
			StudentID:    studentID,
			Status:       models.AssignmentTargetStatusPending,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
	}
	return r.db.Create(&targets).Error
}

func (r *assignmentTargetRepository) MarkAsOverdue() error {
	return r.db.Model(&models.AssignmentTarget{}).
		Joins("JOIN assignments ON assignment_targets.assignment_id = assignments.id").
		Where("assignment_targets.status = ? AND assignments.due_date < ?",
			models.AssignmentTargetStatusPending, time.Now()).
		Update("status", models.AssignmentTargetStatusOverdue).Error
}
