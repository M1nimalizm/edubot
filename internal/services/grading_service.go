package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

type GradingService interface {
	// Feedback CRUD
	CreateFeedback(feedback *models.Feedback) error
	GetFeedback(id uuid.UUID) (*models.Feedback, error)
	UpdateFeedback(feedback *models.Feedback) error
	DeleteFeedback(id uuid.UUID) error

	// Grading operations
	GradeAssignment(assignmentTargetID, teacherID uuid.UUID, score *float64, text string, mediaIDs []uuid.UUID) (*models.Feedback, error)
	GetFeedbacksByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Feedback, error)
	GetFeedbacksByTeacher(teacherID uuid.UUID) ([]*models.Feedback, error)
	GetLatestFeedback(assignmentTargetID uuid.UUID) (*models.Feedback, error)

	// Teacher inbox operations
	GetPendingGrading(teacherID uuid.UUID) ([]*models.AssignmentTarget, error)
	GetGradedAssignments(teacherID uuid.UUID) ([]*models.AssignmentTarget, error)
}

type gradingService struct {
	feedbackRepo         repository.FeedbackRepository
	assignmentTargetRepo repository.AssignmentTargetRepository
	submissionRepo       repository.SubmissionRepository
	userRepo             repository.UserRepository
	notificationRepo     repository.NotificationRepository
	bot                  *telegram.Bot
}

func NewGradingService(
	feedbackRepo repository.FeedbackRepository,
	assignmentTargetRepo repository.AssignmentTargetRepository,
	submissionRepo repository.SubmissionRepository,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
	bot *telegram.Bot,
) GradingService {
	return &gradingService{
		feedbackRepo:         feedbackRepo,
		assignmentTargetRepo: assignmentTargetRepo,
		submissionRepo:       submissionRepo,
		userRepo:             userRepo,
		notificationRepo:     notificationRepo,
		bot:                  bot,
	}
}

func (s *gradingService) CreateFeedback(feedback *models.Feedback) error {
	if feedback.ID == uuid.Nil {
		feedback.ID = uuid.New()
	}
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = time.Now()

	return s.feedbackRepo.Create(feedback)
}

func (s *gradingService) GetFeedback(id uuid.UUID) (*models.Feedback, error) {
	return s.feedbackRepo.GetByID(id)
}

func (s *gradingService) UpdateFeedback(feedback *models.Feedback) error {
	feedback.UpdatedAt = time.Now()
	return s.feedbackRepo.Update(feedback)
}

func (s *gradingService) DeleteFeedback(id uuid.UUID) error {
	return s.feedbackRepo.Delete(id)
}

func (s *gradingService) GradeAssignment(assignmentTargetID, teacherID uuid.UUID, score *float64, text string, mediaIDs []uuid.UUID) (*models.Feedback, error) {
	// Получаем AssignmentTarget
	target, err := s.assignmentTargetRepo.GetByID(assignmentTargetID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что задание отправлено
	if target.Status != models.AssignmentTargetStatusSubmitted {
		return nil, errors.New("assignment not submitted yet")
	}

	// Получаем Assignment для проверки прав учителя
	assignment, err := s.assignmentTargetRepo.GetByID(assignmentTargetID)
	if err != nil {
		return nil, err
	}

	// Проверяем права учителя
	if assignment.Assignment.TeacherID != teacherID {
		return nil, errors.New("access denied: not assignment teacher")
	}

	// Создаем Feedback
	feedback := &models.Feedback{
		ID:                 uuid.New(),
		AssignmentTargetID: assignmentTargetID,
		TeacherID:          teacherID,
		Text:               text,
		Score:              score,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Сохраняем Feedback
	if err := s.CreateFeedback(feedback); err != nil {
		return nil, err
	}

	// Обновляем статус AssignmentTarget
	target.Status = models.AssignmentTargetStatusGraded
	target.Score = score
	target.GradedAt = &feedback.CreatedAt
	target.UpdatedAt = time.Now()

	if err := s.assignmentTargetRepo.Update(target); err != nil {
		return nil, err
	}

	// Обновляем последний Submission с оценкой
	submissions, err := s.submissionRepo.GetByAssignmentTarget(assignmentTargetID)
	if err == nil && len(submissions) > 0 {
		latestSubmission := submissions[0] // Предполагаем, что они отсортированы по дате
		latestSubmission.Grade = s.scoreToString(score)
		latestSubmission.TeacherComments = text
		latestSubmission.Status = "reviewed"
		latestSubmission.ReviewedAt = &feedback.CreatedAt
		latestSubmission.UpdatedAt = time.Now()

		s.submissionRepo.Update(latestSubmission)
	}

	// Отправляем уведомление студенту
	if s.bot != nil {
		student, err := s.userRepo.GetByID(target.StudentID)
		if err == nil && student.TelegramID != 0 {
			s.bot.SendFeedbackNotification(
				student.TelegramID,
				assignment.Assignment.Title,
				assignment.Assignment.Subject,
				s.scoreToString(score),
				text,
			)
		}
	}

	// Создаем уведомление в системе
	s.notificationRepo.Create(&models.Notification{
		UserID:    target.StudentID,
		Type:      models.NotificationTypeGradeReceived,
		Title:     "Получена оценка",
		Message:   "Ваше задание оценено: " + assignment.Assignment.Title,
		Payload:   `{"feedback_id":"` + feedback.ID.String() + `","assignment_target_id":"` + assignmentTargetID.String() + `"}`,
		Channel:   models.NotificationChannelBot,
		Status:    models.NotificationStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	return feedback, nil
}

func (s *gradingService) GetFeedbacksByAssignmentTarget(assignmentTargetID uuid.UUID) ([]*models.Feedback, error) {
	return s.feedbackRepo.GetByAssignmentTarget(assignmentTargetID)
}

func (s *gradingService) GetFeedbacksByTeacher(teacherID uuid.UUID) ([]*models.Feedback, error) {
	return s.feedbackRepo.GetByTeacher(teacherID)
}

func (s *gradingService) GetLatestFeedback(assignmentTargetID uuid.UUID) (*models.Feedback, error) {
	return s.feedbackRepo.GetLatestByAssignmentTarget(assignmentTargetID)
}

func (s *gradingService) GetPendingGrading(teacherID uuid.UUID) ([]*models.AssignmentTarget, error) {
	// Получаем все AssignmentTarget со статусом "submitted" для заданий этого учителя
	return s.assignmentTargetRepo.ListByStatus(models.AssignmentTargetStatusSubmitted)
}

func (s *gradingService) GetGradedAssignments(teacherID uuid.UUID) ([]*models.AssignmentTarget, error) {
	// Получаем все AssignmentTarget со статусом "graded" для заданий этого учителя
	return s.assignmentTargetRepo.ListByStatus(models.AssignmentTargetStatusGraded)
}

// Helper method to convert score to string
func (s *gradingService) scoreToString(score *float64) string {
	if score == nil {
		return "needs_revision"
	}

	switch {
	case *score >= 4.5:
		return "5"
	case *score >= 3.5:
		return "4"
	case *score >= 2.5:
		return "3"
	case *score >= 1.5:
		return "2"
	default:
		return "needs_revision"
	}
}
