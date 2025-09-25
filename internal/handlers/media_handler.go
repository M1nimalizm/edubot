package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/services"
)

// MediaHandler обрабатывает HTTP запросы для медиафайлов
type MediaHandler struct {
	mediaService services.MediaService
}

// NewMediaHandler создает новый обработчик медиафайлов
func NewMediaHandler(mediaService services.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// CreateMediaRequest запрос на создание медиафайла
type CreateMediaRequest struct {
	TelegramFileID   string                `json:"telegram_file_id" binding:"required"`
	TelegramUniqueID string                `json:"telegram_unique_id"`
	ChatID           int64                 `json:"chat_id"`
	MessageID        int                   `json:"message_id"`
	Type             models.MediaType      `json:"type" binding:"required"`
	MimeType         string                `json:"mime_type"`
	Size             int64                 `json:"size"`
	Caption          string                `json:"caption"`
	Scope            models.MediaScope      `json:"scope"`
	EntityType       models.EntityType     `json:"entity_type"`
	EntityID         *uuid.UUID            `json:"entity_id"`
}

// UpdateMediaRequest запрос на обновление медиафайла
type UpdateMediaRequest struct {
	Caption string            `json:"caption"`
	Scope   models.MediaScope `json:"scope"`
}

// GrantAccessRequest запрос на предоставление доступа
type GrantAccessRequest struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	Permission string    `json:"permission" binding:"required"`
}

// CreateMedia создает новый медиафайл
func (h *MediaHandler) CreateMedia(c *gin.Context) {
	var req CreateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	ownerID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	media, err := h.mediaService.CreateMediaFromTelegram(
		req.TelegramFileID,
		req.TelegramUniqueID,
		req.ChatID,
		req.MessageID,
		req.Type,
		req.MimeType,
		req.Size,
		req.Caption,
		ownerID,
		req.Scope,
		req.EntityType,
		req.EntityID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"media": media})
}

// GetMedia получает информацию о медиафайле
func (h *MediaHandler) GetMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	media, err := h.mediaService.GetMediaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"media": media})
}

// StreamMedia стримит медиафайл
func (h *MediaHandler) StreamMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	stream, err := h.mediaService.GetMediaStream(id, userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	// Записываем просмотр
	go func() {
		h.mediaService.RecordView(id, userUUID, 0)
	}()

	// Устанавливаем заголовки для стриминга
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Cache-Control", "public, max-age=3600")

	// Копируем поток в ответ
	io.Copy(c.Writer, stream)
}

// GetThumbnail получает миниатюру медиафайла
func (h *MediaHandler) GetThumbnail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	thumbnail, err := h.mediaService.GetMediaThumbnail(id, userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	defer thumbnail.Close()

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=86400")

	io.Copy(c.Writer, thumbnail)
}

// GetUserMedia получает медиафайлы пользователя
func (h *MediaHandler) GetUserMedia(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	media, err := h.mediaService.GetUserMedia(userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"media": media})
}

// GetPublicMedia получает публичные медиафайлы
func (h *MediaHandler) GetPublicMedia(c *gin.Context) {
	media, err := h.mediaService.GetPublicMedia()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"media": media})
}

// GetEntityMedia получает медиафайлы сущности
func (h *MediaHandler) GetEntityMedia(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityIDStr := c.Param("entity_id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity ID"})
		return
	}

	media, err := h.mediaService.GetEntityMedia(models.EntityType(entityType), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"media": media})
}

// UpdateMedia обновляет медиафайл
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	var req UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	media, err := h.mediaService.GetMediaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	// Проверяем права владельца
	if media.OwnerID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Обновляем поля
	if req.Caption != "" {
		media.Caption = req.Caption
	}
	if req.Scope != "" {
		media.Scope = req.Scope
	}

	err = h.mediaService.UpdateMedia(media)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"media": media})
}

// DeleteMedia удаляет медиафайл
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	err = h.mediaService.DeleteMedia(id, userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "media deleted successfully"})
}

// GetMediaViews получает просмотры медиафайла
func (h *MediaHandler) GetMediaViews(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	views, err := h.mediaService.GetMediaViews(id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"views": views})
}

// GrantAccess предоставляет доступ к медиафайлу
func (h *MediaHandler) GrantAccess(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	var req GrantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	// Проверяем, что пользователь является владельцем медиафайла
	media, err := h.mediaService.GetMediaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	if media.OwnerID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.mediaService.GrantMediaAccess(id, req.UserID, req.Permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access granted successfully"})
}

// RevokeAccess отзывает доступ к медиафайлу
func (h *MediaHandler) RevokeAccess(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	ownerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	ownerUUID, ok := ownerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	// Проверяем, что пользователь является владельцем медиафайла
	media, err := h.mediaService.GetMediaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	if media.OwnerID != ownerUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.mediaService.RevokeMediaAccess(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access revoked successfully"})
}
