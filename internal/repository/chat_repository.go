package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

type ChatRepository interface {
	// Threads
	CreateThread(thread *models.ChatThread) error
	GetThread(id uuid.UUID) (*models.ChatThread, error)
	GetOrCreateStudentTeacherThread(studentID, teacherID uuid.UUID) (*models.ChatThread, error)
	GetOrCreateGroupThread(groupID, teacherID uuid.UUID) (*models.ChatThread, error)
	ListThreadsForUser(userID uuid.UUID) ([]*models.ChatThread, error)
	UpdateThread(thread *models.ChatThread) error
	DeleteThread(id uuid.UUID) error

	// Messages
	CreateMessage(message *models.Message) error
	GetMessage(id uuid.UUID) (*models.Message, error)
	ListMessages(threadID uuid.UUID, limit int, before *time.Time) ([]*models.Message, error)
	UpdateMessage(message *models.Message) error
	DeleteMessage(id uuid.UUID) error

	// Unread counts
	GetUnreadCount(threadID, userID uuid.UUID) (int64, error)
	MarkAsRead(threadID, userID uuid.UUID) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateThread(thread *models.ChatThread) error {
	if thread.ID == uuid.Nil {
		thread.ID = uuid.New()
	}
	return r.db.Create(thread).Error
}

func (r *chatRepository) GetThread(id uuid.UUID) (*models.ChatThread, error) {
	var thread models.ChatThread
	err := r.db.Preload("Student").Preload("Group").Preload("Teacher").
		First(&thread, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *chatRepository) GetOrCreateStudentTeacherThread(studentID, teacherID uuid.UUID) (*models.ChatThread, error) {
	var thread models.ChatThread
	err := r.db.Where("type = ? AND student_id = ? AND teacher_id = ?",
		models.ChatThreadTypeStudentTeacher, studentID, teacherID).
		First(&thread).Error

	if err == gorm.ErrRecordNotFound {
		// Создаем новый тред
		thread = models.ChatThread{
			ID:        uuid.New(),
			Type:      models.ChatThreadTypeStudentTeacher,
			StudentID: &studentID,
			TeacherID: teacherID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := r.CreateThread(&thread); err != nil {
			return nil, err
		}
		return &thread, nil
	}

	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *chatRepository) GetOrCreateGroupThread(groupID, teacherID uuid.UUID) (*models.ChatThread, error) {
	var thread models.ChatThread
	err := r.db.Where("type = ? AND group_id = ? AND teacher_id = ?",
		models.ChatThreadTypeGroup, groupID, teacherID).
		First(&thread).Error

	if err == gorm.ErrRecordNotFound {
		// Создаем новый тред
		thread = models.ChatThread{
			ID:        uuid.New(),
			Type:      models.ChatThreadTypeGroup,
			GroupID:   &groupID,
			TeacherID: teacherID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := r.CreateThread(&thread); err != nil {
			return nil, err
		}
		return &thread, nil
	}

	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *chatRepository) ListThreadsForUser(userID uuid.UUID) ([]*models.ChatThread, error) {
	var threads []*models.ChatThread
	err := r.db.Preload("Student").Preload("Group").Preload("Teacher").
		Where("student_id = ? OR teacher_id = ?", userID, userID).
		Order("last_message_at DESC NULLS LAST, created_at DESC").
		Find(&threads).Error
	return threads, err
}

func (r *chatRepository) UpdateThread(thread *models.ChatThread) error {
	thread.UpdatedAt = time.Now()
	return r.db.Save(thread).Error
}

func (r *chatRepository) DeleteThread(id uuid.UUID) error {
	return r.db.Delete(&models.ChatThread{}, "id = ?", id).Error
}

func (r *chatRepository) CreateMessage(message *models.Message) error {
	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}
	return r.db.Create(message).Error
}

func (r *chatRepository) GetMessage(id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Thread").Preload("Author").Preload("Media").
		First(&message, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *chatRepository) ListMessages(threadID uuid.UUID, limit int, before *time.Time) ([]*models.Message, error) {
	var messages []*models.Message
	query := r.db.Preload("Author").Preload("Media").
		Where("thread_id = ?", threadID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if before != nil {
		query = query.Where("created_at < ?", *before)
	}

	err := query.Find(&messages).Error
	return messages, err
}

func (r *chatRepository) UpdateMessage(message *models.Message) error {
	message.EditedAt = &[]time.Time{time.Now()}[0]
	return r.db.Save(message).Error
}

func (r *chatRepository) DeleteMessage(id uuid.UUID) error {
	return r.db.Delete(&models.Message{}, "id = ?", id).Error
}

func (r *chatRepository) GetUnreadCount(threadID, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Where("thread_id = ? AND author_id != ?", threadID, userID).
		Count(&count).Error
	return count, err
}

func (r *chatRepository) MarkAsRead(threadID, userID uuid.UUID) error {
	// В простой реализации просто возвращаем nil
	// В реальной системе здесь была бы таблица MessageRead
	return nil
}
