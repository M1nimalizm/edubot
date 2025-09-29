package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/repository"
	"edubot/pkg/telegram"
)

type ChatService interface {
	// Thread operations
	GetOrCreateStudentTeacherThread(studentID, teacherID uuid.UUID) (*models.ChatThread, error)
	GetOrCreateGroupThread(groupID, teacherID uuid.UUID) (*models.ChatThread, error)
	GetThread(id uuid.UUID) (*models.ChatThread, error)
	ListThreadsForUser(userID uuid.UUID) ([]*models.ChatThread, error)
	UpdateThread(thread *models.ChatThread) error

	// Message operations
	SendMessage(threadID, authorID uuid.UUID, text *string, mediaIDs []uuid.UUID, kind models.MessageKind) (*models.Message, error)
	GetMessage(id uuid.UUID) (*models.Message, error)
	ListMessages(threadID uuid.UUID, limit int, before *time.Time) ([]*models.Message, error)
	UpdateMessage(message *models.Message) error
	DeleteMessage(id uuid.UUID) error

	// Unread operations
	GetUnreadCount(threadID, userID uuid.UUID) (int64, error)
	MarkAsRead(threadID, userID uuid.UUID) error

	// System messages
	SendSystemMessage(threadID uuid.UUID, text string, kind models.MessageKind) (*models.Message, error)
}

type chatService struct {
	chatRepo         repository.ChatRepository
	userRepo         repository.UserRepository
	groupRepo        repository.GroupRepository
	notificationRepo repository.NotificationRepository
	bot              *telegram.Bot
}

func NewChatService(
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	groupRepo repository.GroupRepository,
	notificationRepo repository.NotificationRepository,
	bot *telegram.Bot,
) ChatService {
	return &chatService{
		chatRepo:         chatRepo,
		userRepo:         userRepo,
		groupRepo:        groupRepo,
		notificationRepo: notificationRepo,
		bot:              bot,
	}
}

func (s *chatService) GetOrCreateStudentTeacherThread(studentID, teacherID uuid.UUID) (*models.ChatThread, error) {
	return s.chatRepo.GetOrCreateStudentTeacherThread(studentID, teacherID)
}

func (s *chatService) GetOrCreateGroupThread(groupID, teacherID uuid.UUID) (*models.ChatThread, error) {
	return s.chatRepo.GetOrCreateGroupThread(groupID, teacherID)
}

func (s *chatService) GetThread(id uuid.UUID) (*models.ChatThread, error) {
	return s.chatRepo.GetThread(id)
}

func (s *chatService) ListThreadsForUser(userID uuid.UUID) ([]*models.ChatThread, error) {
	return s.chatRepo.ListThreadsForUser(userID)
}

func (s *chatService) UpdateThread(thread *models.ChatThread) error {
	thread.UpdatedAt = time.Now()
	return s.chatRepo.UpdateThread(thread)
}

func (s *chatService) SendMessage(threadID, authorID uuid.UUID, text *string, mediaIDs []uuid.UUID, kind models.MessageKind) (*models.Message, error) {
	// Проверяем, что пользователь имеет доступ к треду
	thread, err := s.chatRepo.GetThread(threadID)
	if err != nil {
		return nil, err
	}

	// Проверяем права доступа
	if !s.hasAccessToThread(thread, authorID) {
		return nil, errors.New("access denied to thread")
	}

	// Создаем сообщение
	message := &models.Message{
		ID:        uuid.New(),
		ThreadID:  threadID,
		AuthorID:  authorID,
		Text:      text,
		Kind:      kind,
		CreatedAt: time.Now(),
	}

	// Сохраняем сообщение
	if err := s.chatRepo.CreateMessage(message); err != nil {
		return nil, err
	}

	// Обновляем время последнего сообщения в треде
	thread.LastMessageAt = &message.CreatedAt
	thread.UpdatedAt = time.Now()
	s.chatRepo.UpdateThread(thread)

	// Отправляем уведомления другим участникам треда
	s.notifyThreadParticipants(thread, message)

	return message, nil
}

func (s *chatService) GetMessage(id uuid.UUID) (*models.Message, error) {
	return s.chatRepo.GetMessage(id)
}

func (s *chatService) ListMessages(threadID uuid.UUID, limit int, before *time.Time) ([]*models.Message, error) {
	return s.chatRepo.ListMessages(threadID, limit, before)
}

func (s *chatService) UpdateMessage(message *models.Message) error {
	message.EditedAt = &[]time.Time{time.Now()}[0]
	return s.chatRepo.UpdateMessage(message)
}

func (s *chatService) DeleteMessage(id uuid.UUID) error {
	return s.chatRepo.DeleteMessage(id)
}

func (s *chatService) GetUnreadCount(threadID, userID uuid.UUID) (int64, error) {
	return s.chatRepo.GetUnreadCount(threadID, userID)
}

func (s *chatService) MarkAsRead(threadID, userID uuid.UUID) error {
	return s.chatRepo.MarkAsRead(threadID, userID)
}

func (s *chatService) SendSystemMessage(threadID uuid.UUID, text string, kind models.MessageKind) (*models.Message, error) {
	// Системные сообщения отправляются от имени системы (используем nil authorID)
	systemUserID := uuid.Nil // В реальной системе был бы специальный системный пользователь

	message := &models.Message{
		ID:        uuid.New(),
		ThreadID:  threadID,
		AuthorID:  systemUserID,
		Text:      &text,
		Kind:      kind,
		CreatedAt: time.Now(),
	}

	if err := s.chatRepo.CreateMessage(message); err != nil {
		return nil, err
	}

	// Обновляем время последнего сообщения в треде
	thread, err := s.chatRepo.GetThread(threadID)
	if err == nil {
		thread.LastMessageAt = &message.CreatedAt
		thread.UpdatedAt = time.Now()
		s.chatRepo.UpdateThread(thread)
	}

	return message, nil
}

// Helper method to check if user has access to thread
func (s *chatService) hasAccessToThread(thread *models.ChatThread, userID uuid.UUID) bool {
	switch thread.Type {
	case models.ChatThreadTypeStudentTeacher:
		return (thread.StudentID != nil && *thread.StudentID == userID) || thread.TeacherID == userID
	case models.ChatThreadTypeGroup:
		if thread.TeacherID == userID {
			return true
		}
		if thread.GroupID != nil {
			// Проверяем, является ли пользователь участником группы
			isMember, err := s.groupRepo.IsMember(*thread.GroupID, userID)
			return err == nil && isMember
		}
	}
	return false
}

// Helper method to notify thread participants
func (s *chatService) notifyThreadParticipants(thread *models.ChatThread, message *models.Message) {
	if s.bot == nil {
		return
	}

	var recipientIDs []uuid.UUID

	switch thread.Type {
	case models.ChatThreadTypeStudentTeacher:
		if thread.StudentID != nil && *thread.StudentID != message.AuthorID {
			recipientIDs = append(recipientIDs, *thread.StudentID)
		}
		if thread.TeacherID != message.AuthorID {
			recipientIDs = append(recipientIDs, thread.TeacherID)
		}
	case models.ChatThreadTypeGroup:
		if thread.GroupID != nil {
			members, err := s.groupRepo.ListMembers(*thread.GroupID)
			if err == nil {
				for _, member := range members {
					if member.UserID != message.AuthorID {
						recipientIDs = append(recipientIDs, member.UserID)
					}
				}
			}
		}
	}

	// Отправляем уведомления
	for _, recipientID := range recipientIDs {
		user, err := s.userRepo.GetByID(recipientID)
		if err == nil && user.TelegramID != 0 {
			messageText := "[Медиафайл]"
			if message.Text != nil {
				messageText = *message.Text
			}
			s.bot.SendMessage(user.TelegramID, messageText)
		}

		// Создаем уведомление в системе
		s.notificationRepo.Create(&models.Notification{
			UserID:    recipientID,
			Type:      models.NotificationTypeNewMessage,
			Title:     "Новое сообщение",
			Message:   s.getMessagePreview(message),
			Payload:   `{"thread_id":"` + thread.ID.String() + `","message_id":"` + message.ID.String() + `"}`,
			Channel:   models.NotificationChannelBot,
			Status:    models.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
}

// Helper method to create message preview
func (s *chatService) getMessagePreview(message *models.Message) string {
	if message.Text != nil && len(*message.Text) > 0 {
		text := *message.Text
		if len(text) > 50 {
			return text[:47] + "..."
		}
		return text
	}
	return "[Медиафайл]"
}
