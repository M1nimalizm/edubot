package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

type SubmissionService interface {
	// Submission CRUD
	CreateSubmission(submission *models.Submission) error
	GetSubmission(id uuid.UUID) (*models.Submission, error)
	UpdateSubmission(submission *models.Submission) error
	DeleteSubmission(id uuid.UUID) error

	// Student operations
	SubmitAssignment(assignmentTargetID, studentID uuid.UUID, text *string, mediaIDs []uuid.UUID) (*models.Submission, error)
	GetSubmissionsByStudent(studentID uuid.UUID) ([]*models.Submission, error)
	GetSubmissionsByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Submission, error)

	// Draft operations
	SaveDraft(assignmentTargetID, studentID uuid.UUID, text *string, mediaIDs []uuid.UUID) (*models.Draft, error)
	GetDraft(assignmentTargetID, studentID uuid.UUID) (*models.Draft, error)
	DeleteDraft(assignmentTargetID uuid.UUID) error

	// Status management
	MarkAsLate(submissionID uuid.UUID) error
	GetLateSubmissions() ([]*models.Submission, error)
}

type submissionService struct {
	submissionRepo       repository.SubmissionRepository
	assignmentTargetRepo repository.AssignmentTargetRepository
	draftRepo            repository.DraftRepository
	userRepo             repository.UserRepository
	notificationRepo     repository.NotificationRepository
	bot                  *telegram.Bot
}

func NewSubmissionService(
	submissionRepo repository.SubmissionRepository,
	assignmentTargetRepo repository.AssignmentTargetRepository,
	draftRepo repository.DraftRepository,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
	bot *telegram.Bot,
) SubmissionService {
	return &submissionService{
		submissionRepo:       submissionRepo,
		assignmentTargetRepo: assignmentTargetRepo,
		draftRepo:            draftRepo,
		userRepo:             userRepo,
		notificationRepo:     notificationRepo,
		bot:                  bot,
	}
}

func (s *submissionService) CreateSubmission(submission *models.Submission) error {
	if submission.ID == uuid.Nil {
		submission.ID = uuid.New()
	}
	submission.CreatedAt = time.Now()
	submission.UpdatedAt = time.Now()

	return s.submissionRepo.Create(submission)
}

func (s *submissionService) GetSubmission(id uuid.UUID) (*models.Submission, error) {
	return s.submissionRepo.GetByID(id)
}

func (s *submissionService) UpdateSubmission(submission *models.Submission) error {
	submission.UpdatedAt = time.Now()
	return s.submissionRepo.Update(submission)
}

func (s *submissionService) DeleteSubmission(id uuid.UUID) error {
	return s.submissionRepo.Delete(id)
}

func (s *submissionService) SubmitAssignment(assignmentTargetID, studentID uuid.UUID, text *string, mediaIDs []uuid.UUID) (*models.Submission, error) {
	// Получаем AssignmentTarget
	target, err := s.assignmentTargetRepo.GetByID(assignmentTargetID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что студент имеет право отправлять это задание
	if target.StudentID != studentID {
		return nil, errors.New("assignment not assigned to this student")
	}

	// Проверяем статус задания
	if target.Status == models.AssignmentTargetStatusGraded {
		return nil, errors.New("assignment already graded")
	}

	// Получаем Assignment для проверки дедлайна
	assignment, err := s.assignmentTargetRepo.GetByID(assignmentTargetID)
	if err != nil {
		return nil, err
	}

	// Определяем, просрочено ли задание
	isLate := time.Now().After(assignment.Assignment.DueDate)

	// Создаем Submission
	submission := &models.Submission{
		ID:                 uuid.New(),
		AssignmentID:       assignment.Assignment.ID,
		AssignmentTargetID: &assignmentTargetID,
		UserID:             studentID,
		Text:               text,
		Status:             "submitted",
		SubmittedAt:        time.Now(),
		IsLate:             isLate,
		Attempt:            1, // TODO: подсчитать количество попыток
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Сохраняем Submission
	if err := s.CreateSubmission(submission); err != nil {
		return nil, err
	}

	// Обновляем статус AssignmentTarget
	target.Status = models.AssignmentTargetStatusSubmitted
	target.SubmittedAt = &submission.SubmittedAt
	if isLate {
		target.IsLate = true
	}
	target.UpdatedAt = time.Now()

	if err := s.assignmentTargetRepo.Update(target); err != nil {
		return nil, err
	}

	// Удаляем черновик, если он был
	s.draftRepo.DeleteByAssignmentTarget(assignmentTargetID)

	// Отправляем уведомление учителю
	if s.bot != nil {
		teacher, err := s.userRepo.GetByID(assignment.Assignment.TeacherID)
		if err == nil && teacher.TelegramID != 0 {
			s.bot.SendAssignmentCompletedNotification(
				teacher.TelegramID,
				assignment.Assignment.Title,
				assignment.Assignment.Subject,
				assignment.Assignment.DueDate.Format("2006-01-02 15:04"),
			)
		}
	}

	// Создаем уведомление в системе
	s.notificationRepo.Create(&models.Notification{
		UserID:    assignment.Assignment.TeacherID,
		Type:      models.NotificationTypeNewAssignment,
		Title:     "Новая отправка задания",
		Message:   "Студент отправил задание: " + assignment.Assignment.Title,
		Payload:   `{"submission_id":"` + submission.ID.String() + `","assignment_target_id":"` + assignmentTargetID.String() + `"}`,
		Channel:   models.NotificationChannelBot,
		Status:    models.NotificationStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	return submission, nil
}

func (s *submissionService) GetSubmissionsByStudent(studentID uuid.UUID) ([]*models.Submission, error) {
	return s.submissionRepo.GetByStudentID(studentID)
}

func (s *submissionService) GetSubmissionsByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Submission, error) {
	return s.submissionRepo.GetByAssignmentTarget(assignmentTargetID)
}

func (s *submissionService) SaveDraft(assignmentTargetID, studentID uuid.UUID, text *string, mediaIDs []uuid.UUID) (*models.Draft, error) {
	// Проверяем, существует ли уже черновик
	existingDraft, err := s.draftRepo.GetByAssignmentTarget(assignmentTargetID)
	if err == nil && existingDraft != nil {
		// Обновляем существующий черновик
		existingDraft.Text = text
		existingDraft.MediaIDs = s.serializeMediaIDs(mediaIDs)
		existingDraft.UpdatedAt = time.Now()

		if err := s.draftRepo.Update(existingDraft); err != nil {
			return nil, err
		}
		return existingDraft, nil
	}

	// Создаем новый черновик
	draft := &models.Draft{
		ID:                 uuid.New(),
		AssignmentTargetID: &assignmentTargetID,
		StudentID:          studentID,
		Text:               text,
		MediaIDs:           s.serializeMediaIDs(mediaIDs),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.draftRepo.Create(draft); err != nil {
		return nil, err
	}

	return draft, nil
}

func (s *submissionService) GetDraft(assignmentTargetID, studentID uuid.UUID) (*models.Draft, error) {
	return s.draftRepo.GetByAssignmentTarget(assignmentTargetID)
}

func (s *submissionService) DeleteDraft(assignmentTargetID uuid.UUID) error {
	return s.draftRepo.DeleteByAssignmentTarget(assignmentTargetID)
}

func (s *submissionService) MarkAsLate(submissionID uuid.UUID) error {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return err
	}

	submission.IsLate = true
	submission.UpdatedAt = time.Now()

	return s.submissionRepo.Update(submission)
}

func (s *submissionService) GetLateSubmissions() ([]*models.Submission, error) {
	return s.submissionRepo.GetLateSubmissions()
}

// Helper method to serialize media IDs to JSON string
func (s *submissionService) serializeMediaIDs(mediaIDs []uuid.UUID) string {
	if len(mediaIDs) == 0 {
		return "[]"
	}

	result := "["
	for i, id := range mediaIDs {
		if i > 0 {
			result += ","
		}
		result += `"` + id.String() + `"`
	}
	result += "]"

	return result
}
