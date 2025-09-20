package services

import (
	"fmt"
	"time"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/storage"
	"edubot/pkg/telegram"
	"mime/multipart"

	"github.com/google/uuid"
)

// AssignmentService представляет сервис для работы с заданиями
type AssignmentService struct {
	assignmentRepo *repository.AssignmentRepository
	submissionRepo *repository.SubmissionRepository
	attachmentRepo *repository.AttachmentRepository
	userRepo       *repository.UserRepository
	storage        *storage.Storage
	telegramBot    *telegram.Bot
}

// NewAssignmentService создает новый сервис заданий
func NewAssignmentService(
	assignmentRepo *repository.AssignmentRepository,
	submissionRepo *repository.SubmissionRepository,
	attachmentRepo *repository.AttachmentRepository,
	userRepo *repository.UserRepository,
	storage *storage.Storage,
	telegramBot *telegram.Bot,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
		attachmentRepo: attachmentRepo,
		userRepo:       userRepo,
		storage:        storage,
		telegramBot:    telegramBot,
	}
}

// CreateAssignmentRequest представляет запрос на создание задания
type CreateAssignmentRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Subject     string      `json:"subject"`
	Deadline    time.Time   `json:"deadline"`
	UserIDs     []uuid.UUID `json:"user_ids"`
}

// CreateAssignment создает новое задание
func (s *AssignmentService) CreateAssignment(req *CreateAssignmentRequest, creatorID uuid.UUID) (*models.Assignment, error) {
	// Создаем задание
	assignment := &models.Assignment{
		Title:       req.Title,
		Description: req.Description,
		Subject:     req.Subject,
		Deadline:    req.Deadline,
		CreatedBy:   creatorID,
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	// Назначаем задание пользователям
	for _, userID := range req.UserIDs {
		if err := s.assignmentRepo.AssignToUser(assignment.ID, userID); err != nil {
			// Логируем ошибку, но продолжаем
			fmt.Printf("Failed to assign assignment to user %s: %v\n", userID, err)
		}
	}

	// Отправляем уведомления пользователям
	s.notifyUsersAboutAssignment(assignment, req.UserIDs)

	return assignment, nil
}

// GetAssignmentsForUser получает задания пользователя
func (s *AssignmentService) GetAssignmentsForUser(userID uuid.UUID) ([]models.Assignment, error) {
	return s.assignmentRepo.ListByUser(userID)
}

// GetAssignmentDetails получает детали задания
func (s *AssignmentService) GetAssignmentDetails(assignmentID uuid.UUID, userID uuid.UUID) (*models.Assignment, error) {
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	// Загружаем решение пользователя если оно есть
	submission, err := s.submissionRepo.GetByAssignmentAndUser(assignmentID, userID)
	if err == nil {
		assignment.Submissions = []models.Submission{*submission}
	}

	return assignment, nil
}

// SubmitSolution загружает решение задания
func (s *AssignmentService) SubmitSolution(assignmentID, userID uuid.UUID, files []*multipart.FileHeader) (*models.Submission, error) {
	// Проверяем, существует ли задание
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	// Проверяем дедлайн
	if time.Now().After(assignment.Deadline) {
		return nil, fmt.Errorf("assignment deadline has passed")
	}

	// Проверяем, не сдано ли уже решение
	existingSubmission, err := s.submissionRepo.GetByAssignmentAndUser(assignmentID, userID)
	if err == nil && existingSubmission != nil {
		return nil, fmt.Errorf("solution already submitted")
	}

	// Создаем новое решение
	submission := &models.Submission{
		AssignmentID: assignmentID,
		UserID:       userID,
		Status:       "submitted",
		SubmittedAt:  time.Now(),
	}

	if err := s.submissionRepo.Create(submission); err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	// Сохраняем файлы
	for _, file := range files {
		filePath, err := s.storage.SaveFile(file, userID, "submissions")
		if err != nil {
			return nil, fmt.Errorf("failed to save file %s: %w", file.Filename, err)
		}

		attachment := &models.Attachment{
			FileName:     file.Filename,
			OriginalName: file.Filename,
			FilePath:     filePath,
			FileSize:     file.Size,
			MimeType:     file.Header.Get("Content-Type"),
			SubmissionID: &submission.ID,
		}

		if err := s.attachmentRepo.Create(attachment); err != nil {
			return nil, fmt.Errorf("failed to create attachment: %w", err)
		}
	}

	return submission, nil
}

// GradeSubmission оценивает решение
func (s *AssignmentService) GradeSubmission(submissionID uuid.UUID, grade int, comments string) error {
	if grade < 1 || grade > 5 {
		return fmt.Errorf("grade must be between 1 and 5")
	}

	if err := s.submissionRepo.Grade(submissionID, grade, comments); err != nil {
		return fmt.Errorf("failed to grade submission: %w", err)
	}

	// Получаем решение для отправки уведомления
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	// Отправляем уведомление ученику
	if err := s.telegramBot.SendGradeNotification(
		submission.User.TelegramID,
		submission.Assignment.Title,
		grade,
		comments,
	); err != nil {
		fmt.Printf("Failed to send grade notification: %v\n", err)
	}

	return nil
}

// GetPendingSubmissions получает непроверенные решения
func (s *AssignmentService) GetPendingSubmissions() ([]models.Submission, error) {
	return s.submissionRepo.ListPending()
}

// GetUpcomingDeadlines получает задания с приближающимися дедлайнами
func (s *AssignmentService) GetUpcomingDeadlines(hours int) ([]models.Assignment, error) {
	return s.assignmentRepo.ListUpcoming(hours)
}

// SendDeadlineReminders отправляет напоминания о дедлайнах
func (s *AssignmentService) SendDeadlineReminders() error {
	assignments, err := s.GetUpcomingDeadlines(24) // За 24 часа
	if err != nil {
		return fmt.Errorf("failed to get upcoming deadlines: %w", err)
	}

	for _, assignment := range assignments {
		hoursLeft := int(time.Until(assignment.Deadline).Hours())

		// Отправляем напоминания всем назначенным пользователям
		for _, userAssignment := range assignment.UserAssignments {
			if err := s.telegramBot.SendDeadlineReminder(
				userAssignment.User.TelegramID,
				assignment.Title,
				hoursLeft,
			); err != nil {
				fmt.Printf("Failed to send deadline reminder to user %d: %v\n",
					userAssignment.User.TelegramID, err)
			}
		}
	}

	return nil
}

// notifyUsersAboutAssignment отправляет уведомления о новом задании
func (s *AssignmentService) notifyUsersAboutAssignment(assignment *models.Assignment, userIDs []uuid.UUID) {
	for _, userID := range userIDs {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			fmt.Printf("Failed to get user %s: %v\n", userID, err)
			continue
		}

		if err := s.telegramBot.SendAssignmentNotification(
			user.TelegramID,
			assignment.Title,
			assignment.Subject,
			assignment.Deadline.Format("02.01.2006 15:04"),
		); err != nil {
			fmt.Printf("Failed to send assignment notification to user %d: %v\n",
				user.TelegramID, err)
		}
	}
}
