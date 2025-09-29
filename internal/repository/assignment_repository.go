package repository

import (
	"edubot/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssignmentRepository interface {
	Create(assignment *models.Assignment) error
	GetByID(id uuid.UUID) (*models.Assignment, error)
	GetByStudentID(studentID uuid.UUID) ([]models.Assignment, error)
	GetByTeacherID(teacherID uuid.UUID) ([]models.Assignment, error)
	GetUpcomingDeadlines(studentID uuid.UUID, days int) ([]models.Assignment, error)
	Update(assignment *models.Assignment) error
	UpdateStatus(id uuid.UUID, status string) error
	MarkCompleted(id uuid.UUID) error
	Delete(id uuid.UUID) error

	// Comment CRUD
	CreateComment(comment *models.Comment) error
	GetCommentsByAssignmentID(assignmentID uuid.UUID) ([]models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id uuid.UUID) error

	// Content CRUD
	CreateContent(content *models.Content) error
	GetContentByID(id uuid.UUID) (*models.Content, error)
	GetContentBySubject(subject string, grade int) ([]models.Content, error)
	GetContentByTeacherID(teacherID uuid.UUID) ([]models.Content, error)
	UpdateContent(content *models.Content) error
	DeleteContent(id uuid.UUID) error

	// Student Progress
	CreateProgress(progress *models.StudentProgress) error
	GetProgressByStudentID(studentID uuid.UUID) ([]models.StudentProgress, error)
	UpdateProgress(progress *models.StudentProgress) error
	GetProgressBySubject(studentID uuid.UUID, subject string) (*models.StudentProgress, error)

	// Submission methods
	CreateSubmission(submission *models.Submission) error
	GetSubmissionByID(id uuid.UUID) (*models.Submission, error)
	GetSubmissionsByAssignmentID(assignmentID uuid.UUID) ([]models.Submission, error)
	GetSubmissionsByUserID(userID uuid.UUID) ([]models.Submission, error)
	UpdateSubmission(submission *models.Submission) error
	DeleteSubmission(id uuid.UUID) error

	// Group methods
	GetByGroupID(groupID uuid.UUID) ([]*models.Assignment, error)
}

type assignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

// Assignment CRUD
func (r *assignmentRepository) Create(assignment *models.Assignment) error {
	return r.db.Create(assignment).Error
}

func (r *assignmentRepository) GetByID(id uuid.UUID) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.Preload("Teacher").Preload("Student").Preload("Comments.Author").
		Where("id = ?", id).First(&assignment).Error
	return &assignment, err
}

func (r *assignmentRepository) GetByStudentID(studentID uuid.UUID) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Preload("Teacher").Preload("Comments.Author").
		Where("student_id = ?", studentID).
		Order("due_date ASC").Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) GetByTeacherID(teacherID uuid.UUID) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Preload("Student").Preload("Comments.Author").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) GetUpcomingDeadlines(studentID uuid.UUID, days int) ([]models.Assignment, error) {
	var assignments []models.Assignment
	deadline := time.Now().AddDate(0, 0, days)

	err := r.db.Preload("Teacher").
		Where("student_id = ? AND due_date <= ? AND status != 'completed'", studentID, deadline).
		Order("due_date ASC").Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) Update(assignment *models.Assignment) error {
	return r.db.Save(assignment).Error
}

func (r *assignmentRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Assignment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *assignmentRepository) MarkCompleted(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Assignment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": &now,
		}).Error
}

func (r *assignmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Assignment{}, id).Error
}

// Comment CRUD
func (r *assignmentRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *assignmentRepository) GetCommentsByAssignmentID(assignmentID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Preload("Author").
		Where("assignment_id = ?", assignmentID).
		Order("created_at ASC").Find(&comments).Error
	return comments, err
}

func (r *assignmentRepository) UpdateComment(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *assignmentRepository) DeleteComment(id uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

// Content CRUD
func (r *assignmentRepository) CreateContent(content *models.Content) error {
	return r.db.Create(content).Error
}

func (r *assignmentRepository) GetContentByID(id uuid.UUID) (*models.Content, error) {
	var content models.Content
	err := r.db.Preload("Creator").
		Where("id = ?", id).First(&content).Error
	return &content, err
}

func (r *assignmentRepository) GetContentBySubject(subject string, grade int) ([]models.Content, error) {
	var content []models.Content
	err := r.db.Preload("Creator").
		Where("subject = ? AND grade = ?", subject, grade).
		Order("created_at DESC").Find(&content).Error
	return content, err
}

func (r *assignmentRepository) GetContentByTeacherID(teacherID uuid.UUID) ([]models.Content, error) {
	var content []models.Content
	err := r.db.Preload("Creator").
		Where("created_by = ?", teacherID).
		Order("created_at DESC").Find(&content).Error
	return content, err
}

func (r *assignmentRepository) UpdateContent(content *models.Content) error {
	return r.db.Save(content).Error
}

func (r *assignmentRepository) DeleteContent(id uuid.UUID) error {
	return r.db.Delete(&models.Content{}, id).Error
}

// Student Progress
func (r *assignmentRepository) CreateProgress(progress *models.StudentProgress) error {
	return r.db.Create(progress).Error
}

func (r *assignmentRepository) GetProgressByStudentID(studentID uuid.UUID) ([]models.StudentProgress, error) {
	var progress []models.StudentProgress
	err := r.db.Preload("Student").
		Where("student_id = ?", studentID).
		Find(&progress).Error
	return progress, err
}

func (r *assignmentRepository) UpdateProgress(progress *models.StudentProgress) error {
	return r.db.Save(progress).Error
}

func (r *assignmentRepository) GetProgressBySubject(studentID uuid.UUID, subject string) (*models.StudentProgress, error) {
	var progress models.StudentProgress
	err := r.db.Preload("Student").
		Where("student_id = ? AND subject = ?", studentID, subject).
		First(&progress).Error
	return &progress, err
}

// Submission methods
func (r *assignmentRepository) CreateSubmission(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

func (r *assignmentRepository) GetSubmissionByID(id uuid.UUID) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Preload("Assignment").Preload("User").
		Where("id = ?", id).First(&submission).Error
	return &submission, err
}

func (r *assignmentRepository) GetSubmissionsByAssignmentID(assignmentID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	err := r.db.Preload("User").
		Where("assignment_id = ?", assignmentID).
		Order("submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}

func (r *assignmentRepository) GetSubmissionsByUserID(userID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	err := r.db.Preload("Assignment").
		Where("user_id = ?", userID).
		Order("submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}

func (r *assignmentRepository) UpdateSubmission(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

func (r *assignmentRepository) DeleteSubmission(id uuid.UUID) error {
	return r.db.Delete(&models.Submission{}, id).Error
}

// Group methods
func (r *assignmentRepository) GetByGroupID(groupID uuid.UUID) ([]*models.Assignment, error) {
	var assignments []*models.Assignment
	err := r.db.Preload("Teacher").Preload("Group").
		Where("group_id = ?", groupID).
		Order("created_at DESC").Find(&assignments).Error
	return assignments, err
}
