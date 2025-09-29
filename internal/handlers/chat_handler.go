package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"edubot/internal/models"
	"edubot/internal/services"
)

type ChatHandler struct {
	chatService services.ChatService
}

func NewChatHandler(chatService services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// GET /api/chat/threads - Получить список чатов пользователя
func (h *ChatHandler) GetThreads(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем треды пользователя
	threads, err := h.chatService.ListThreadsForUser(userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get threads"})
		return
	}

	// Добавляем информацию о непрочитанных сообщениях
	var result []gin.H
	for _, thread := range threads {
		unreadCount, _ := h.chatService.GetUnreadCount(thread.ID, userUUID)

		result = append(result, gin.H{
			"id":              thread.ID,
			"type":            thread.Type,
			"student":         thread.Student,
			"group":           thread.Group,
			"teacher":         thread.Teacher,
			"last_message_at": thread.LastMessageAt,
			"unread_count":    unreadCount,
			"created_at":      thread.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"threads": result,
	})
}

// GET /api/chat/threads/:id - Получить детали чата
func (h *ChatHandler) GetThread(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	threadIDStr := c.Param("id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Получаем тред
	thread, err := h.chatService.GetThread(threadID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	// Получаем количество непрочитанных сообщений
	unreadCount, _ := h.chatService.GetUnreadCount(threadID, userUUID)

	c.JSON(http.StatusOK, gin.H{
		"thread":       thread,
		"unread_count": unreadCount,
	})
}

// GET /api/chat/threads/:id/messages - Получить сообщения чата
func (h *ChatHandler) GetMessages(c *gin.Context) {
	threadIDStr := c.Param("id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Получаем параметры запроса
	limitStr := c.DefaultQuery("limit", "50")
	beforeStr := c.Query("before")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	var before *time.Time
	if beforeStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			before = &parsedTime
		}
	}

	// Получаем сообщения
	messages, err := h.chatService.ListMessages(threadID, limit, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"before":   before,
	})
}

// POST /api/chat/threads/:id/messages - Отправить сообщение
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	threadIDStr := c.Param("id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	var request struct {
		Text     *string     `json:"text"`
		MediaIDs []uuid.UUID `json:"media_ids"`
		Kind     string      `json:"kind,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Определяем тип сообщения
	kind := models.MessageKindMessage
	if request.Kind != "" {
		kind = models.MessageKind(request.Kind)
	}

	// Отправляем сообщение
	message, err := h.chatService.SendMessage(threadID, userUUID, request.Text, request.MediaIDs, kind)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// PUT /api/chat/messages/:id - Редактировать сообщение
func (h *ChatHandler) UpdateMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var request struct {
		Text *string `json:"text"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Получаем сообщение
	message, err := h.chatService.GetMessage(messageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Проверяем, что пользователь является автором сообщения
	if message.AuthorID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Обновляем сообщение
	message.Text = request.Text
	if err := h.chatService.UpdateMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// DELETE /api/chat/messages/:id - Удалить сообщение
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Получаем сообщение
	message, err := h.chatService.GetMessage(messageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Проверяем, что пользователь является автором сообщения
	if message.AuthorID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Удаляем сообщение
	if err := h.chatService.DeleteMessage(messageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
	})
}

// POST /api/chat/threads/:id/read - Отметить чат как прочитанный
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	threadIDStr := c.Param("id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Отмечаем как прочитанное
	if err := h.chatService.MarkAsRead(threadID, userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thread marked as read",
	})
}

// POST /api/chat/threads/student-teacher - Создать или получить чат с учителем
func (h *ChatHandler) GetOrCreateStudentTeacherThread(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request struct {
		TeacherID uuid.UUID `json:"teacher_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Создаем или получаем тред
	thread, err := h.chatService.GetOrCreateStudentTeacherThread(userUUID, request.TeacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"thread": thread,
	})
}

// POST /api/chat/threads/group - Создать или получить групповой чат
func (h *ChatHandler) GetOrCreateGroupThread(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request struct {
		GroupID uuid.UUID `json:"group_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Создаем или получаем тред
	thread, err := h.chatService.GetOrCreateGroupThread(request.GroupID, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"thread": thread,
	})
}
