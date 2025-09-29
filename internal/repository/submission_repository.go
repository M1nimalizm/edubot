package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

type SubmissionRepository interface {
	Create(submission *models.Submission) error
	GetByID(id uuid.UUID) (*models.Submission, error)
	GetByStudentID(studentID uuid.UUID) ([]*models.Submission, error)
	GetByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Submission, error)
	Update(submission *models.Submission) error
	Delete(id uuid.UUID) error
	GetLateSubmissions() ([]*models.Submission, error)
}

type submissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) SubmissionRepository {
	return &submissionRepository{db: db}
}

func (r *submissionRepository) Create(submission *models.Submission) error {
	if submission.ID == uuid.Nil {
		submission.ID = uuid.New()
	}
	return r.db.Create(submission).Error
}

func (r *submissionRepository) GetByID(id uuid.UUID) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Preload("Assignment").Preload("AssignmentTarget").Preload("User").Preload("Files").
		First(&submission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *submissionRepository) GetByStudentID(studentID uuid.UUID) ([]*models.Submission, error) {
	var submissions []*models.Submission
	err := r.db.Preload("Assignment").Preload("AssignmentTarget").Preload("Files").
		Where("user_id = ?", studentID).
		Order("submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}

func (r *submissionRepository) GetByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Submission, error) {
	var submissions []*models.Submission
	err := r.db.Preload("Assignment").Preload("User").Preload("Files").
		Where("assignment_target_id = ?", assignmentTargetID).
		Order("submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}

func (r *submissionRepository) Update(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

func (r *submissionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Submission{}, "id = ?", id).Error
}

func (r *submissionRepository) GetLateSubmissions() ([]*models.Submission, error) {
	var submissions []*models.Submission
	err := r.db.Preload("Assignment").Preload("User").
		Where("is_late = ?", true).
		Order("submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}
