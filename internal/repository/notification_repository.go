package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"edubot/internal/models"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByID(id uuid.UUID) (*models.Notification, error)
	ListByUser(userID uuid.UUID) ([]*models.Notification, error)
	ListByStatus(status models.NotificationStatus) ([]*models.Notification, error)
	ListByChannel(channel models.NotificationChannel) ([]*models.Notification, error)
	Update(notification *models.Notification) error
	Delete(id uuid.UUID) error

	// Методы для массовых операций
	CreateForGroup(userIDs []uuid.UUID, notificationType models.NotificationType, title, message, payload string) error
	MarkAsSent(id uuid.UUID) error
	MarkAsRead(id uuid.UUID) error
	CleanupOld(olderThan time.Time) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByID(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.Preload("User").First(&notification, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) ListByUser(userID uuid.UUID) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) ListByStatus(status models.NotificationStatus) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Preload("User").
		Where("status = ?", status).
		Order("created_at ASC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) ListByChannel(channel models.NotificationChannel) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Preload("User").
		Where("channel = ? AND status = ?", channel, models.NotificationStatusPending).
		Order("created_at ASC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) Update(notification *models.Notification) error {
	return r.db.Save(notification).Error
}

func (r *notificationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Notification{}, "id = ?", id).Error
}

func (r *notificationRepository) CreateForGroup(userIDs []uuid.UUID, notificationType models.NotificationType, title, message, payload string) error {
	var notifications []models.Notification
	for _, userID := range userIDs {
		notifications = append(notifications, models.Notification{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      notificationType,
			Title:     title,
			Message:   message,
			Payload:   payload,
			Channel:   models.NotificationChannelBot, // По умолчанию через бота
			Status:    models.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	return r.db.Create(&notifications).Error
}

func (r *notificationRepository) MarkAsSent(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     models.NotificationStatusSent,
			"sent_at":    &now,
			"updated_at": now,
		}).Error
}

func (r *notificationRepository) MarkAsRead(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     models.NotificationStatusRead,
			"read_at":    &now,
			"updated_at": now,
		}).Error
}

func (r *notificationRepository) CleanupOld(olderThan time.Time) error {
	return r.db.Where("created_at < ? AND status IN (?)",
		olderThan,
		[]models.NotificationStatus{models.NotificationStatusSent, models.NotificationStatusRead}).
		Delete(&models.Notification{}).Error
}
