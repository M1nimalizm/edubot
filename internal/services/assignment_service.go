package services

import (
	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
	"errors"
	"time"

	"github.com/google/uuid"
)

type LegacyAssignmentService struct {
	assignmentRepo repository.AssignmentRepository
	userRepo       repository.UserRepository
	mediaService   MediaService
	telegramBot    *telegram.Bot
}

func NewLegacyAssignmentService(assignmentRepo repository.AssignmentRepository, userRepo repository.UserRepository, mediaService MediaService, telegramBot *telegram.Bot) *LegacyAssignmentService {
	return &LegacyAssignmentService{
		assignmentRepo: assignmentRepo,
		userRepo:       userRepo,
		mediaService:   mediaService,
		telegramBot:    telegramBot,
	}
}

// Assignment methods
func (s *LegacyAssignmentService) CreateAssignment(assignment *models.Assignment) error {
	assignment.CreatedAt = time.Now()
	if err := s.assignmentRepo.Create(assignment); err != nil {
		return err
	}

	// Загружаем ученика для уведомления
	if assignment.StudentID != nil {
		student, err := s.userRepo.GetByID(*assignment.StudentID)
		if err == nil && student.TelegramID != 0 {
			s.telegramBot.SendAssignmentNotification(student.TelegramID, assignment.Title, assignment.Subject, assignment.DueDate.Format("02.01.2006 15:04"))
		}
	}
	return nil
}

func (s *LegacyAssignmentService) GetAssignmentByID(id uuid.UUID) (*models.Assignment, error) {
	return s.assignmentRepo.GetByID(id)
}

func (s *LegacyAssignmentService) GetAssignmentsByStudentID(studentID uuid.UUID) ([]models.Assignment, error) {
	return s.assignmentRepo.GetByStudentID(studentID)
}

func (s *LegacyAssignmentService) GetAssignmentsByTeacherID(teacherID uuid.UUID) ([]models.Assignment, error) {
	return s.assignmentRepo.GetByTeacherID(teacherID)
}

func (s *LegacyAssignmentService) GetUpcomingDeadlines(studentID uuid.UUID, days int) ([]models.Assignment, error) {
	return s.assignmentRepo.GetUpcomingDeadlines(studentID, days)
}

func (s *LegacyAssignmentService) UpdateAssignment(assignment *models.Assignment) error {
	return s.assignmentRepo.Update(assignment)
}

func (s *LegacyAssignmentService) MarkAssignmentCompleted(assignmentID uuid.UUID, studentID uuid.UUID) error {
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return err
	}
	if assignment.StudentID != nil && *assignment.StudentID != studentID {
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

func (s *LegacyAssignmentService) DeleteAssignment(assignmentID uuid.UUID, teacherID uuid.UUID) error {
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
func (s *LegacyAssignmentService) AddComment(comment *models.Comment) error {
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
		if assignment.StudentID != nil {
			student, err := s.userRepo.GetByID(*assignment.StudentID)
			if err == nil {
				recipientTelegramID = student.TelegramID
			}
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

func (s *LegacyAssignmentService) GetCommentsByAssignmentID(assignmentID uuid.UUID) ([]models.Comment, error) {
	return s.assignmentRepo.GetCommentsByAssignmentID(assignmentID)
}

func (s *LegacyAssignmentService) UpdateComment(comment *models.Comment) error {
	return s.assignmentRepo.UpdateComment(comment)
}

func (s *LegacyAssignmentService) DeleteComment(commentID uuid.UUID) error {
	return s.assignmentRepo.DeleteComment(commentID)
}

// Content methods
func (s *LegacyAssignmentService) CreateContent(content *models.Content) error {
	content.CreatedAt = time.Now()
	return s.assignmentRepo.CreateContent(content)
}

func (s *LegacyAssignmentService) GetContentByID(id uuid.UUID) (*models.Content, error) {
	return s.assignmentRepo.GetContentByID(id)
}

func (s *LegacyAssignmentService) GetContentBySubject(subject string, grade int) ([]models.Content, error) {
	return s.assignmentRepo.GetContentBySubject(subject, grade)
}

func (s *LegacyAssignmentService) GetContentByTeacherID(teacherID uuid.UUID) ([]models.Content, error) {
	return s.assignmentRepo.GetContentByTeacherID(teacherID)
}

func (s *LegacyAssignmentService) UpdateContent(content *models.Content) error {
	return s.assignmentRepo.UpdateContent(content)
}

func (s *LegacyAssignmentService) DeleteContent(contentID uuid.UUID, teacherID uuid.UUID) error {
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
func (s *LegacyAssignmentService) GetStudentProgress(studentID uuid.UUID) ([]models.StudentProgress, error) {
	return s.assignmentRepo.GetProgressByStudentID(studentID)
}

func (s *LegacyAssignmentService) GetStudentProgressBySubject(studentID uuid.UUID, subject string) (*models.StudentProgress, error) {
	return s.assignmentRepo.GetProgressBySubject(studentID, subject)
}

func (s *LegacyAssignmentService) updateStudentProgress(studentID uuid.UUID, subject string) {
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

// AddAssignmentMedia добавляет медиафайл к заданию
func (s *LegacyAssignmentService) AddAssignmentMedia(assignmentID uuid.UUID, mediaID uuid.UUID, userID uuid.UUID) error {
	// Проверяем, что пользователь является создателем задания
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return err
	}

	if assignment.CreatedBy != userID {
		return errors.New("access denied: not assignment creator")
	}

	// Получаем медиафайл
	media, err := s.mediaService.GetMediaByID(mediaID)
	if err != nil {
		return err
	}

	// Обновляем медиафайл, привязывая его к заданию
	media.EntityType = models.EntityTypeAssignment
	media.EntityID = &assignmentID
	media.Scope = models.MediaScopeStudent // Медиа задания доступно ученикам

	return s.mediaService.UpdateMedia(media)
}

// GetAssignmentMedia получает медиафайлы задания
func (s *LegacyAssignmentService) GetAssignmentMedia(assignmentID uuid.UUID, userID uuid.UUID) ([]*models.Media, error) {
	// Проверяем доступ к заданию
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Проверяем права доступа
	if (assignment.StudentID != nil && *assignment.StudentID != userID) && assignment.TeacherID != userID && user.Role != models.RoleTeacher {
		return nil, errors.New("access denied")
	}

	// Получаем медиафайлы задания
	return s.mediaService.GetEntityMedia(models.EntityTypeAssignment, assignmentID)
}

// SubmitAssignmentWithMedia создает submission с медиафайлами
func (s *LegacyAssignmentService) SubmitAssignmentWithMedia(assignmentID uuid.UUID, userID uuid.UUID, mediaIDs []uuid.UUID, comments string) error {
	// Проверяем, что пользователь является учеником
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.Role != models.RoleStudent {
		return errors.New("only students can submit assignments")
	}

	// Проверяем задание
	assignment, err := s.assignmentRepo.GetByID(assignmentID)
	if err != nil {
		return err
	}

	if assignment.StudentID != nil && *assignment.StudentID != userID {
		return errors.New("assignment not assigned to this student")
	}

	// Создаем submission
	submission := &models.Submission{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		UserID:       userID,
		Status:       "submitted",
		Text:         &comments,
		SubmittedAt:  time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Сохраняем submission
	if err := s.assignmentRepo.CreateSubmission(submission); err != nil {
		return err
	}

	// Привязываем медиафайлы к submission
	for _, mediaID := range mediaIDs {
		media, err := s.mediaService.GetMediaByID(mediaID)
		if err != nil {
			continue // Пропускаем несуществующие медиафайлы
		}

		media.EntityType = models.EntityTypeSubmission
		media.EntityID = &submission.ID
		media.Scope = models.MediaScopeTeacher // Медиа submission доступно учителю

		if err := s.mediaService.UpdateMedia(media); err != nil {
			continue // Пропускаем ошибки обновления
		}
	}

	// Обновляем статус задания
	assignment.Status = "completed"
	assignment.UpdatedAt = time.Now()

	if err := s.assignmentRepo.Update(assignment); err != nil {
		return err
	}

	// Отправляем уведомление учителю
	teacher, err := s.userRepo.GetByID(assignment.TeacherID)
	if err == nil && teacher.TelegramID != 0 {
		s.telegramBot.SendAssignmentCompletedNotification(
			teacher.TelegramID,
			assignment.Title,
			assignment.Subject,
			assignment.DueDate.Format("2006-01-02 15:04"),
		)
	}

	return nil
}

// GetSubmissionMedia получает медиафайлы submission
func (s *LegacyAssignmentService) GetSubmissionMedia(submissionID uuid.UUID, userID uuid.UUID) ([]*models.Media, error) {
	// Получаем submission
	submission, err := s.assignmentRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}

	// Получаем задание
	assignment, err := s.assignmentRepo.GetByID(submission.AssignmentID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Проверяем права доступа
	if submission.UserID != userID && assignment.TeacherID != userID && user.Role != models.RoleTeacher {
		return nil, errors.New("access denied")
	}

	// Получаем медиафайлы submission
	return s.mediaService.GetEntityMedia(models.EntityTypeSubmission, submissionID)
}

// AddFeedbackMedia добавляет медиафайл с фидбэком к submission
func (s *LegacyAssignmentService) AddFeedbackMedia(submissionID uuid.UUID, mediaID uuid.UUID, userID uuid.UUID) error {
	// Получаем submission
	submission, err := s.assignmentRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return err
	}

	// Получаем задание
	assignment, err := s.assignmentRepo.GetByID(submission.AssignmentID)
	if err != nil {
		return err
	}

	// Проверяем, что пользователь является учителем
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.Role != models.RoleTeacher || assignment.TeacherID != userID {
		return errors.New("access denied: only assignment teacher can add feedback")
	}

	// Получаем медиафайл
	media, err := s.mediaService.GetMediaByID(mediaID)
	if err != nil {
		return err
	}

	// Обновляем медиафайл, привязывая его к submission как фидбэк
	media.EntityType = models.EntityTypeReview
	media.EntityID = &submissionID
	media.Scope = models.MediaScopeStudent // Фидбэк доступен ученику

	return s.mediaService.UpdateMedia(media)
}

// GetFeedbackMedia получает медиафайлы с фидбэком
func (s *LegacyAssignmentService) GetFeedbackMedia(submissionID uuid.UUID, userID uuid.UUID) ([]*models.Media, error) {
	// Получаем submission
	submission, err := s.assignmentRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}

	// Получаем задание
	assignment, err := s.assignmentRepo.GetByID(submission.AssignmentID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Проверяем права доступа
	if submission.UserID != userID && assignment.TeacherID != userID && user.Role != models.RoleTeacher {
		return nil, errors.New("access denied")
	}

	// Получаем медиафайлы фидбэка
	return s.mediaService.GetEntityMedia(models.EntityTypeReview, submissionID)
}

// GetTeacherSubmissions получает все submissions для учителя
func (s *LegacyAssignmentService) GetTeacherSubmissions(teacherID uuid.UUID) ([]*models.Submission, error) {
	teacher, err := s.userRepo.GetByID(teacherID)
	if err != nil {
		return nil, err
	}
	if teacher.Role != models.RoleTeacher {
		return nil, errors.New("access denied: user is not a teacher")
	}

	// Получаем все задания учителя
	assignments, err := s.assignmentRepo.GetByTeacherID(teacherID)
	if err != nil {
		return nil, err
	}

	var allSubmissions []*models.Submission
	for _, assignment := range assignments {
		submissions, err := s.assignmentRepo.GetSubmissionsByAssignmentID(assignment.ID)
		if err != nil {
			continue
		}

		// Конвертируем в указатели
		for i := range submissions {
			allSubmissions = append(allSubmissions, &submissions[i])
		}
	}

	return allSubmissions, nil
}

// SubmitTeacherFeedback добавляет фидбэк учителя к submission
func (s *LegacyAssignmentService) SubmitTeacherFeedback(submissionID uuid.UUID, teacherID uuid.UUID, comments string, grade string) error {
	submission, err := s.assignmentRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return err
	}

	assignment, err := s.assignmentRepo.GetByID(submission.AssignmentID)
	if err != nil {
		return err
	}

	teacher, err := s.userRepo.GetByID(teacherID)
	if err != nil {
		return err
	}

	if teacher.Role != models.RoleTeacher || assignment.TeacherID != teacherID {
		return errors.New("access denied: only assignment teacher can provide feedback")
	}

	// Обновляем submission с фидбэком
	submission.TeacherComments = comments
	submission.Grade = grade
	submission.Status = "reviewed"
	if grade == "needs_revision" {
		submission.Status = "needs_revision"
	}
	submission.ReviewedAt = &time.Time{}
	*submission.ReviewedAt = time.Now()
	submission.UpdatedAt = time.Now()

	if err := s.assignmentRepo.UpdateSubmission(submission); err != nil {
		return err
	}

	// Отправляем уведомление ученику
	student, err := s.userRepo.GetByID(submission.UserID)
	if err == nil && student.TelegramID != 0 {
		s.telegramBot.SendFeedbackNotification(
			student.TelegramID,
			assignment.Title,
			assignment.Subject,
			grade,
			comments,
		)
	}

	return nil
}
