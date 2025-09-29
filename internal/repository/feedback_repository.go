package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

type FeedbackRepository interface {
	Create(feedback *models.Feedback) error
	GetByID(id uuid.UUID) (*models.Feedback, error)
	GetByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Feedback, error)
	GetByTeacher(teacherID uuid.UUID) ([]*models.Feedback, error)
	Update(feedback *models.Feedback) error
	Delete(id uuid.UUID) error
	GetLatestByAssignmentTarget(assignmentTargetID uuid.UUID) (*models.Feedback, error)
}

type feedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

func (r *feedbackRepository) Create(feedback *models.Feedback) error {
	if feedback.ID == uuid.Nil {
		feedback.ID = uuid.New()
	}
	return r.db.Create(feedback).Error
}

func (r *feedbackRepository) GetByID(id uuid.UUID) (*models.Feedback, error) {
	var feedback models.Feedback
	err := r.db.Preload("AssignmentTarget").Preload("AssignmentTarget.Assignment").
		Preload("AssignmentTarget.Student").Preload("Teacher").Preload("Media").
		First(&feedback, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *feedbackRepository) GetByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Feedback, error) {
	var feedbacks []*models.Feedback
	err := r.db.Preload("Teacher").Preload("Media").
		Where("assignment_target_id = ?", assignmentTargetID).
		Order("created_at DESC").
		Find(&feedbacks).Error
	return feedbacks, err
}

func (r *feedbackRepository) GetByTeacher(teacherID uuid.UUID) ([]*models.Feedback, error) {
	var feedbacks []*models.Feedback
	err := r.db.Preload("AssignmentTarget").Preload("AssignmentTarget.Assignment").
		Preload("AssignmentTarget.Student").Preload("Media").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").
		Find(&feedbacks).Error
	return feedbacks, err
}

func (r *feedbackRepository) Update(feedback *models.Feedback) error {
	return r.db.Save(feedback).Error
}

func (r *feedbackRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Feedback{}, "id = ?", id).Error
}

func (r *feedbackRepository) GetLatestByAssignmentTarget(assignmentTargetID uuid.UUID) (*models.Feedback, error) {
	var feedback models.Feedback
	err := r.db.Preload("Teacher").Preload("Media").
		Where("assignment_target_id = ?", assignmentTargetID).
		Order("created_at DESC").
		First(&feedback).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}
