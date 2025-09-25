package services

import (
	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AssignmentService struct {
	assignmentRepo *repository.AssignmentRepository
	userRepo       repository.UserRepository
	telegramBot    *telegram.Bot
}

func NewAssignmentService(assignmentRepo *repository.AssignmentRepository, userRepo repository.UserRepository, telegramBot *telegram.Bot) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		userRepo:       userRepo,
		telegramBot:    telegramBot,
	}
}

// Assignment methods
func (s *AssignmentService) CreateAssignment(assignment *models.Assignment) error {
	assignment.CreatedAt = time.Now()
	if err := s.assignmentRepo.Create(assignment); err != nil {
		return err
	}

	// Загружаем ученика для уведомления
	student, err := s.userRepo.GetByID(assignment.StudentID)
	if err == nil && student.TelegramID != 0 {
		s.telegramBot.SendAssignmentNotification(student.TelegramID, assignment.Title, assignment.Subject, assignment.DueDate.Format("02.01.2006 15:04"))
	}
	return nil
}

func (s *AssignmentService) GetAssignmentByID(id uuid.UUID) (*models.Assignment, error) {
	return s.assignmentRepo.GetByID(id)
}

func (s *AssignmentService) GetAssignmentsByStudentID(studentID uuid.UUID) ([]models.Assignment, error) {
	return s.assignmentRepo.GetByStudentID(studentID)
}

func (s *AssignmentService) GetAssignmentsByTeacherID(teacherID uuid.UUID) ([]models.Assignment, error) {
	return s.assignmentRepo.GetByTeacherID(teacherID)
}

func (s *AssignmentService) GetUpcomingDeadlines(studentID uuid.UUID, days int) ([]models.Assignment, error) {
	return s.assignmentRepo.GetUpcomingDeadlines(studentID, days)
}

func (s *AssignmentService) UpdateAssignment(assignment *models.Assignment) error {
	return s.assignmentRepo.Update(assignment)
}

func (s *AssignmentService) MarkAssignmentCompleted(assignmentID uuid.UUID, studentID uuid.UUID) error {
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return err
	}
	if assignment.StudentID != studentID {
		return errors.New("assignment does not belong to student")
	}
	if err := s.assignmentRepo.MarkCompleted(assignmentID); err != nil {
		return err
	}

	teacher, err := s.userRepo.GetByID(assignment.TeacherID)
	if err == nil && teacher.TelegramID != 0 {
		s.telegramBot.SendAssignmentCompletedNotification(teacher.TelegramID, assignment.Title, assignment.Subject, assignment.DueDate.Format("02.01.2006 15:04"))
	}

	s.updateStudentProgress(studentID, assignment.Subject)
	return nil
}

func (s *AssignmentService) DeleteAssignment(assignmentID uuid.UUID, teacherID uuid.UUID) error {
	// Проверяем, что задание принадлежит учителю
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return err
	}

	if assignment.TeacherID != teacherID {
		return errors.New("assignment does not belong to teacher")
	}

	return s.assignmentRepo.Delete(assignmentID)
}

// Comment methods
func (s *AssignmentService) AddComment(comment *models.Comment) error {
	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()
	if err := s.assignmentRepo.CreateComment(comment); err != nil {
		return err
	}

	assignment, err := s.assignmentRepo.GetByID(comment.AssignmentID)
	if err != nil {
		return err
	}

	var recipientTelegramID int64
	if comment.AuthorType == "teacher" {
		student, err := s.userRepo.GetByID(assignment.StudentID)
		if err == nil {
			recipientTelegramID = student.TelegramID
		}
	} else {
		teacher, err := s.userRepo.GetByID(assignment.TeacherID)
		if err == nil {
			recipientTelegramID = teacher.TelegramID
		}
	}

	if recipientTelegramID != 0 {
		s.telegramBot.SendCommentNotification(recipientTelegramID, comment.Content, assignment.Title, assignment.Subject)
	}
	return nil
}

func (s *AssignmentService) GetCommentsByAssignmentID(assignmentID uuid.UUID) ([]models.Comment, error) {
	return s.assignmentRepo.GetCommentsByAssignmentID(assignmentID)
}

func (s *AssignmentService) UpdateComment(comment *models.Comment) error {
	return s.assignmentRepo.UpdateComment(comment)
}

func (s *AssignmentService) DeleteComment(commentID uuid.UUID) error {
	return s.assignmentRepo.DeleteComment(commentID)
}

// Content methods
func (s *AssignmentService) CreateContent(content *models.Content) error {
	content.CreatedAt = time.Now()
	return s.assignmentRepo.CreateContent(content)
}

func (s *AssignmentService) GetContentByID(id uuid.UUID) (*models.Content, error) {
	return s.assignmentRepo.GetContentByID(id)
}

func (s *AssignmentService) GetContentBySubject(subject string, grade int) ([]models.Content, error) {
	return s.assignmentRepo.GetContentBySubject(subject, grade)
}

func (s *AssignmentService) GetContentByTeacherID(teacherID uuid.UUID) ([]models.Content, error) {
	return s.assignmentRepo.GetContentByTeacherID(teacherID)
}

func (s *AssignmentService) UpdateContent(content *models.Content) error {
	return s.assignmentRepo.UpdateContent(content)
}

func (s *AssignmentService) DeleteContent(contentID uuid.UUID, teacherID uuid.UUID) error {
	// Проверяем, что контент принадлежит учителю
	content, err := s.assignmentRepo.GetContentByID(contentID)
	if err != nil {
		return err
	}

	if content.CreatedBy != teacherID {
		return errors.New("content does not belong to teacher")
	}

	return s.assignmentRepo.DeleteContent(contentID)
}

// Progress methods
func (s *AssignmentService) GetStudentProgress(studentID uuid.UUID) ([]models.StudentProgress, error) {
	return s.assignmentRepo.GetProgressByStudentID(studentID)
}

func (s *AssignmentService) GetStudentProgressBySubject(studentID uuid.UUID, subject string) (*models.StudentProgress, error) {
	return s.assignmentRepo.GetProgressBySubject(studentID, subject)
}

func (s *AssignmentService) updateStudentProgress(studentID uuid.UUID, subject string) {
	// Получаем или создаем прогресс
	progress, err := s.assignmentRepo.GetProgressBySubject(studentID, subject)
	if err != nil {
		// Создаем новый прогресс
		progress = &models.StudentProgress{
			ID:        uuid.New(),
			StudentID: studentID,
			Subject:   subject,
			Level:     1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		s.assignmentRepo.CreateProgress(progress)
	}

	// Подсчитываем статистику
	assignments, err := s.assignmentRepo.GetByStudentID(studentID)
	if err != nil {
		return
	}

	var completed, total int
	for _, assignment := range assignments {
		if assignment.Subject == subject {
			total++
			if assignment.Status == "completed" {
				completed++
			}
		}
	}

	// Обновляем прогресс
	progress.CompletedAssignments = completed
	progress.TotalAssignments = total
	progress.LastActivity = time.Now()
	progress.UpdatedAt = time.Now()

	if total > 0 {
		progress.AverageScore = float64(completed) / float64(total) * 100
	}

	s.assignmentRepo.UpdateProgress(progress)
}
