package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

type DraftRepository interface {
	Create(draft *models.Draft) error
	GetByID(id uuid.UUID) (*models.Draft, error)
	GetByStudent(studentID uuid.UUID) ([]*models.Draft, error)
	GetByAssignmentTarget(assignmentTargetID uuid.UUID) (*models.Draft, error)
	Update(draft *models.Draft) error
	Delete(id uuid.UUID) error
	DeleteByAssignmentTarget(assignmentTargetID uuid.UUID) error
}

type draftRepository struct {
	db *gorm.DB
}

func NewDraftRepository(db *gorm.DB) DraftRepository {
	return &draftRepository{db: db}
}

func (r *draftRepository) Create(draft *models.Draft) error {
	if draft.ID == uuid.Nil {
		draft.ID = uuid.New()
	}
	return r.db.Create(draft).Error
}

func (r *draftRepository) GetByID(id uuid.UUID) (*models.Draft, error) {
	var draft models.Draft
	err := r.db.Preload("AssignmentTarget").Preload("Student").
		First(&draft, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *draftRepository) GetByStudent(studentID uuid.UUID) ([]*models.Draft, error) {
	var drafts []*models.Draft
	err := r.db.Preload("AssignmentTarget").Preload("AssignmentTarget.Assignment").
		Where("student_id = ?", studentID).
		Order("updated_at DESC").
		Find(&drafts).Error
	return drafts, err
}

func (r *draftRepository) GetByAssignmentTarget(assignmentTargetID uuid.UUID) (*models.Draft, error) {
	var draft models.Draft
	err := r.db.Preload("AssignmentTarget").Preload("Student").
		Where("assignment_target_id = ?", assignmentTargetID).
		First(&draft).Error
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *draftRepository) Update(draft *models.Draft) error {
	draft.UpdatedAt = time.Now()
	return r.db.Save(draft).Error
}

func (r *draftRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Draft{}, "id = ?", id).Error
}

func (r *draftRepository) DeleteByAssignmentTarget(assignmentTargetID uuid.UUID) error {
	return r.db.Where("assignment_target_id = ?", assignmentTargetID).
		Delete(&models.Draft{}).Error
}
