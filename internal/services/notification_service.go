package services

import (
	"log"
	"time"

	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

type NotificationService interface {
	// Notification CRUD
	CreateNotification(notification *models.Notification) error
	GetNotification(id uuid.UUID) (*models.Notification, error)
	UpdateNotification(notification *models.Notification) error
	DeleteNotification(id uuid.UUID) error

	// User notifications
	ListNotificationsByUser(userID uuid.UUID) ([]*models.Notification, error)
	MarkAsRead(notificationID uuid.UUID) error
	MarkAsSent(notificationID uuid.UUID) error

	// Batch operations
	CreateForGroup(userIDs []uuid.UUID, notificationType models.NotificationType, title, message, payload string) error
	SendPendingNotifications() error

	// Deadline management
	ScheduleDeadlineReminders() error
	ScheduleOverdueNotifications() error
	CleanupOldNotifications() error
}

type notificationService struct {
	notificationRepo     repository.NotificationRepository
	assignmentTargetRepo repository.AssignmentTargetRepository
	assignmentRepo       repository.AssignmentRepository
	userRepo             repository.UserRepository
	bot                  *telegram.Bot
}

func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	assignmentTargetRepo repository.AssignmentTargetRepository,
	assignmentRepo repository.AssignmentRepository,
	userRepo repository.UserRepository,
	bot *telegram.Bot,
) NotificationService {
	return &notificationService{
		notificationRepo:     notificationRepo,
		assignmentTargetRepo: assignmentTargetRepo,
		assignmentRepo:       assignmentRepo,
		userRepo:             userRepo,
		bot:                  bot,
	}
}

func (s *notificationService) CreateNotification(notification *models.Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	return s.notificationRepo.Create(notification)
}

func (s *notificationService) GetNotification(id uuid.UUID) (*models.Notification, error) {
	return s.notificationRepo.GetByID(id)
}

func (s *notificationService) UpdateNotification(notification *models.Notification) error {
	return s.notificationRepo.Update(notification)
}

func (s *notificationService) DeleteNotification(id uuid.UUID) error {
	return s.notificationRepo.Delete(id)
}

func (s *notificationService) ListNotificationsByUser(userID uuid.UUID) ([]*models.Notification, error) {
	return s.notificationRepo.ListByUser(userID)
}

func (s *notificationService) MarkAsRead(notificationID uuid.UUID) error {
	return s.notificationRepo.MarkAsRead(notificationID)
}

func (s *notificationService) MarkAsSent(notificationID uuid.UUID) error {
	return s.notificationRepo.MarkAsSent(notificationID)
}

func (s *notificationService) CreateForGroup(userIDs []uuid.UUID, notificationType models.NotificationType, title, message, payload string) error {
	return s.notificationRepo.CreateForGroup(userIDs, notificationType, title, message, payload)
}

func (s *notificationService) SendPendingNotifications() error {
	// Получаем все ожидающие уведомления
	notifications, err := s.notificationRepo.ListByStatus(models.NotificationStatusPending)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := s.sendNotification(notification); err != nil {
			log.Printf("Failed to send notification %s: %v", notification.ID, err)
			continue
		}

		// Помечаем как отправленное
		s.notificationRepo.MarkAsSent(notification.ID)
	}

	return nil
}

func (s *notificationService) ScheduleDeadlineReminders() error {
	// Получаем все активные задания (пока используем простую реализацию)
	// В реальной системе здесь был бы метод GetActiveAssignments()
	// Пока пропускаем эту функциональность
	return nil
}

func (s *notificationService) ScheduleOverdueNotifications() error {
	// Получаем просроченные задания
	overdueTargets, err := s.assignmentTargetRepo.ListByStatus(models.AssignmentTargetStatusOverdue)
	if err != nil {
		return err
	}

	for _, target := range overdueTargets {
		// Создаем уведомление о просрочке
		s.notificationRepo.Create(&models.Notification{
			UserID:    target.StudentID,
			Type:      models.NotificationTypeOverdue,
			Title:     "Задание просрочено",
			Message:   "Ваше задание просрочено: " + target.Assignment.Title,
			Payload:   `{"assignment_target_id":"` + target.ID.String() + `"}`,
			Channel:   models.NotificationChannelBot,
			Status:    models.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return nil
}

func (s *notificationService) CleanupOldNotifications() error {
	// Удаляем уведомления старше 30 дней
	cutoffDate := time.Now().Add(-30 * 24 * time.Hour)
	return s.notificationRepo.CleanupOld(cutoffDate)
}

// Helper method to send notification via appropriate channel
func (s *notificationService) sendNotification(notification *models.Notification) error {
	switch notification.Channel {
	case models.NotificationChannelBot:
		if s.bot != nil {
			user, err := s.userRepo.GetByID(notification.UserID)
			if err == nil && user.TelegramID != 0 {
				message := notification.Title + "\n\n" + notification.Message
				return s.bot.SendMessage(user.TelegramID, message)
			}
		}
	case models.NotificationChannelInApp:
		// In-app уведомления обрабатываются на фронтенде
		return nil
	case models.NotificationChannelEmail:
		// Email уведомления - пока не реализованы
		return nil
	}

	return nil
}

// Helper method to create deadline reminder
func (s *notificationService) createDeadlineReminder(assignment *models.Assignment, timeRemaining string) {
	// Получаем студентов, которым назначено задание
	var studentIDs []uuid.UUID

	if assignment.GroupID != nil {
		// Групповое задание - получаем участников группы
		members, err := s.assignmentTargetRepo.ListByAssignment(assignment.ID)
		if err == nil {
			for _, target := range members {
				studentIDs = append(studentIDs, target.StudentID)
			}
		}
	} else if assignment.StudentID != nil {
		// Индивидуальное задание
		studentIDs = append(studentIDs, *assignment.StudentID)
	}

	// Создаем уведомления для всех студентов
	for _, studentID := range studentIDs {
		s.notificationRepo.Create(&models.Notification{
			UserID:    studentID,
			Type:      models.NotificationTypeDeadlineReminder,
			Title:     "Напоминание о дедлайне",
			Message:   "До сдачи задания осталось " + timeRemaining + ": " + assignment.Title,
			Payload:   `{"assignment_id":"` + assignment.ID.String() + `"}`,
			Channel:   models.NotificationChannelBot,
			Status:    models.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
}
