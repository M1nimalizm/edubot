package repository

import (
	"time"

	"edubot/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssignmentRepository представляет репозиторий для работы с заданиями
type AssignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository создает новый репозиторий заданий
func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Create создает новое задание
func (r *AssignmentRepository) Create(assignment *models.Assignment) error {
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	return r.db.Create(assignment).Error
}

// GetByID получает задание по ID
func (r *AssignmentRepository) GetByID(id uuid.UUID) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.Preload("Creator").Preload("Attachments").Preload("UserAssignments.User").
		First(&assignment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// List получает список всех заданий
func (r *AssignmentRepository) List() ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Preload("Creator").Preload("Attachments").Order("created_at DESC").Find(&assignments).Error
	return assignments, err
}

// ListBySubject получает задания по предмету
func (r *AssignmentRepository) ListBySubject(subject string) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Where("subject = ?", subject).Preload("Creator").Preload("Attachments").
		Order("created_at DESC").Find(&assignments).Error
	return assignments, err
}

// ListByUser получает задания пользователя
func (r *AssignmentRepository) ListByUser(userID uuid.UUID) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Joins("JOIN user_assignments ON assignments.id = user_assignments.assignment_id").
		Where("user_assignments.user_id = ?", userID).
		Preload("Creator").Preload("Attachments").Preload("Submissions", "user_id = ?", userID).
		Order("assignments.created_at DESC").Find(&assignments).Error
	return assignments, err
}

// ListUpcoming получает задания с приближающимися дедлайнами
func (r *AssignmentRepository) ListUpcoming(hours int) ([]models.Assignment, error) {
	var assignments []models.Assignment
	deadline := time.Now().Add(time.Duration(hours) * time.Hour)

	err := r.db.Where("deadline <= ? AND deadline > ?", deadline, time.Now()).
		Preload("Creator").Preload("UserAssignments.User").Find(&assignments).Error
	return assignments, err
}

// Update обновляет задание
func (r *AssignmentRepository) Update(assignment *models.Assignment) error {
	return r.db.Save(assignment).Error
}

// Delete удаляет задание
func (r *AssignmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Assignment{}, "id = ?", id).Error
}

// AssignToUser назначает задание пользователю
func (r *AssignmentRepository) AssignToUser(assignmentID, userID uuid.UUID) error {
	userAssignment := models.UserAssignment{
		ID:           uuid.New(),
		UserID:       userID,
		AssignmentID: assignmentID,
		AssignedAt:   time.Now(),
	}
	return r.db.Create(&userAssignment).Error
}

// UnassignFromUser снимает назначение задания с пользователя
func (r *AssignmentRepository) UnassignFromUser(assignmentID, userID uuid.UUID) error {
	return r.db.Where("assignment_id = ? AND user_id = ?", assignmentID, userID).
		Delete(&models.UserAssignment{}).Error
}

// SubmissionRepository представляет репозиторий для работы с решениями
type SubmissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository создает новый репозиторий решений
func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

// Create создает новое решение
func (r *SubmissionRepository) Create(submission *models.Submission) error {
	if submission.ID == uuid.Nil {
		submission.ID = uuid.New()
	}
	return r.db.Create(submission).Error
}

// GetByID получает решение по ID
func (r *SubmissionRepository) GetByID(id uuid.UUID) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Preload("Assignment").Preload("User").Preload("Files").
		First(&submission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetByAssignmentAndUser получает решение по заданию и пользователю
func (r *SubmissionRepository) GetByAssignmentAndUser(assignmentID, userID uuid.UUID) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Where("assignment_id = ? AND user_id = ?", assignmentID, userID).
		Preload("Assignment").Preload("User").Preload("Files").First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// ListByAssignment получает все решения по заданию
func (r *SubmissionRepository) ListByAssignment(assignmentID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	err := r.db.Where("assignment_id = ?", assignmentID).
		Preload("User").Preload("Files").Order("submitted_at DESC").Find(&submissions).Error
	return submissions, err
}

// ListByUser получает все решения пользователя
func (r *SubmissionRepository) ListByUser(userID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	err := r.db.Where("user_id = ?", userID).
		Preload("Assignment").Preload("Files").Order("submitted_at DESC").Find(&submissions).Error
	return submissions, err
}

// ListPending получает непроверенные решения
func (r *SubmissionRepository) ListPending() ([]models.Submission, error) {
	var submissions []models.Submission
	err := r.db.Where("status = ?", "submitted").
		Preload("Assignment").Preload("User").Preload("Files").
		Order("submitted_at ASC").Find(&submissions).Error
	return submissions, err
}

// Update обновляет решение
func (r *SubmissionRepository) Update(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

// Grade оценивает решение
func (r *SubmissionRepository) Grade(id uuid.UUID, grade int, comments string) error {
	now := time.Now()
	return r.db.Model(&models.Submission{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    "graded",
			"grade":     grade,
			"comments":  comments,
			"graded_at": &now,
		}).Error
}

// Delete удаляет решение
func (r *SubmissionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Submission{}, "id = ?", id).Error
}
