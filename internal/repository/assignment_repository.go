package repository

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"edubot/internal/models"
)

type AssignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Assignment CRUD
func (r *AssignmentRepository) Create(assignment *models.Assignment) error {
	return r.db.Create(assignment).Error
}

func (r *AssignmentRepository) GetByID(id uuid.UUID) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.Preload("Teacher").Preload("Student").Preload("Comments.Author").
		Where("id = ?", id).First(&assignment).Error
	return &assignment, err
}

func (r *AssignmentRepository) GetByStudentID(studentID uuid.UUID) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Preload("Teacher").Preload("Comments.Author").
		Where("student_id = ?", studentID).
		Order("due_date ASC").Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) GetByTeacherID(teacherID uuid.UUID) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Preload("Student").Preload("Comments.Author").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) GetUpcomingDeadlines(studentID uuid.UUID, days int) ([]models.Assignment, error) {
	var assignments []models.Assignment
	deadline := time.Now().AddDate(0, 0, days)
	
	err := r.db.Preload("Teacher").
		Where("student_id = ? AND due_date <= ? AND status != 'completed'", studentID, deadline).
		Order("due_date ASC").Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) Update(assignment *models.Assignment) error {
	return r.db.Save(assignment).Error
}

func (r *AssignmentRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Assignment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *AssignmentRepository) MarkCompleted(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Assignment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": "completed",
			"completed_at": &now,
		}).Error
}

func (r *AssignmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Assignment{}, id).Error
}

// Comment CRUD
func (r *AssignmentRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *AssignmentRepository) GetCommentsByAssignmentID(assignmentID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Preload("Author").
		Where("assignment_id = ?", assignmentID).
		Order("created_at ASC").Find(&comments).Error
	return comments, err
}

func (r *AssignmentRepository) UpdateComment(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *AssignmentRepository) DeleteComment(id uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

// Content CRUD
func (r *AssignmentRepository) CreateContent(content *models.Content) error {
	return r.db.Create(content).Error
}

func (r *AssignmentRepository) GetContentByID(id uuid.UUID) (*models.Content, error) {
	var content models.Content
	err := r.db.Preload("Teacher").
		Where("id = ?", id).First(&content).Error
	return &content, err
}

func (r *AssignmentRepository) GetContentBySubject(subject string, grade int) ([]models.Content, error) {
	var content []models.Content
	err := r.db.Preload("Teacher").
		Where("subject = ? AND grade = ?", subject, grade).
		Order("created_at DESC").Find(&content).Error
	return content, err
}

func (r *AssignmentRepository) GetContentByTeacherID(teacherID uuid.UUID) ([]models.Content, error) {
	var content []models.Content
	err := r.db.Preload("Teacher").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").Find(&content).Error
	return content, err
}

func (r *AssignmentRepository) UpdateContent(content *models.Content) error {
	return r.db.Save(content).Error
}

func (r *AssignmentRepository) DeleteContent(id uuid.UUID) error {
	return r.db.Delete(&models.Content{}, id).Error
}

// Student Progress
func (r *AssignmentRepository) CreateProgress(progress *models.StudentProgress) error {
	return r.db.Create(progress).Error
}

func (r *AssignmentRepository) GetProgressByStudentID(studentID uuid.UUID) ([]models.StudentProgress, error) {
	var progress []models.StudentProgress
	err := r.db.Preload("Student").
		Where("student_id = ?", studentID).
		Find(&progress).Error
	return progress, err
}

func (r *AssignmentRepository) UpdateProgress(progress *models.StudentProgress) error {
	return r.db.Save(progress).Error
}

func (r *AssignmentRepository) GetProgressBySubject(studentID uuid.UUID, subject string) (*models.StudentProgress, error) {
	var progress models.StudentProgress
	err := r.db.Preload("Student").
		Where("student_id = ? AND subject = ?", studentID, subject).
		First(&progress).Error
	return &progress, err
}