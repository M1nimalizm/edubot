package services

import (
	"errors"
	"time"
	"github.com/google/uuid"
	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

type AssignmentService struct {
	assignmentRepo *repository.AssignmentRepository
	userRepo       *repository.UserRepository
	telegramBot    *telegram.Bot
}

func NewAssignmentService(assignmentRepo *repository.AssignmentRepository, userRepo *repository.UserRepository, telegramBot *telegram.Bot) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		userRepo:       userRepo,
		telegramBot:    telegramBot,
	}
}

// Assignment methods
func (s *AssignmentService) CreateAssignment(assignment *models.Assignment) error {
	// Устанавливаем статус по умолчанию
	assignment.Status = "assigned"
	assignment.CreatedAt = time.Now()
	
	// Создаем задание
	if err := s.assignmentRepo.Create(assignment); err != nil {
		return err
	}
	
	// Уведомляем ученика через Telegram
	if assignment.Student.TelegramID != 0 {
		s.telegramBot.SendAssignmentNotification(assignment.Student.TelegramID, assignment)
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
	// Проверяем, что задание принадлежит ученику
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return err
	}
	
	if assignment.StudentID != studentID {
		return errors.New("assignment does not belong to student")
	}
	
	// Отмечаем как выполненное
	if err := s.assignmentRepo.MarkCompleted(assignmentID); err != nil {
		return err
	}
	
	// Уведомляем учителя
	if assignment.Teacher.TelegramID != 0 {
		s.telegramBot.SendAssignmentCompletedNotification(assignment.Teacher.TelegramID, assignment)
	}
	
	// Обновляем прогресс ученика
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
	comment.CreatedAt = time.Now()
	
	if err := s.assignmentRepo.CreateComment(comment); err != nil {
		return err
	}
	
	// Уведомляем получателя комментария
	assignment, err := s.assignmentRepo.GetByID(comment.AssignmentID)
	if err != nil {
		return err
	}
	
	var recipientTelegramID int64
	if comment.AuthorType == "teacher" {
		recipientTelegramID = assignment.Student.TelegramID
	} else {
		recipientTelegramID = assignment.Teacher.TelegramID
	}
	
	if recipientTelegramID != 0 {
		s.telegramBot.SendCommentNotification(recipientTelegramID, comment, assignment)
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
	
	if content.TeacherID != teacherID {
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