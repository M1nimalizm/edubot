package handlers

import (
	"net/http"

	"edubot/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContentHandler представляет обработчик образовательного контента
type ContentHandler struct {
	contentService *services.ContentService
}

// NewContentHandler создает новый обработчик контента
func NewContentHandler(contentService *services.ContentService) *ContentHandler {
	return &ContentHandler{
		contentService: contentService,
	}
}


// CreateContent создает новый контент (только для преподавателя)
func (h *ContentHandler) CreateContent(c *gin.Context) {
	var req CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID создателя из контекста
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	creatorID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем файлы из формы
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided"})
		return
	}

	content, err := h.contentService.CreateContent(&services.CreateContentRequest{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Category:    req.Category,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
	}, creatorID, files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, content)
}

// GetContent получает контент по ID
func (h *ContentHandler) GetContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	content, err := h.contentService.GetContent(contentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Отмечаем как просмотренное если пользователь авторизован
	if userIDStr, exists := c.Get("user_id"); exists {
		if userID, err := uuid.Parse(userIDStr.(string)); err == nil {
			h.contentService.MarkAsViewed(contentID, userID)
		}
	}

	c.JSON(http.StatusOK, content)
}

// ListContent получает список всего публичного контента
func (h *ContentHandler) ListContent(c *gin.Context) {
	content, err := h.contentService.ListContent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

// ListContentByType получает контент по типу
func (h *ContentHandler) ListContentByType(c *gin.Context) {
	contentType := c.Param("type")
	content, err := h.contentService.ListContentByType(contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

// ListContentByCategory получает контент по категории
func (h *ContentHandler) ListContentByCategory(c *gin.Context) {
	category := c.Param("category")
	content, err := h.contentService.ListContentByCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

// SearchContent выполняет поиск контента
func (h *ContentHandler) SearchContent(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	content, err := h.contentService.SearchContent(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

// GetViewedContent получает просмотренный пользователем контент
func (h *ContentHandler) GetViewedContent(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	content, err := h.contentService.GetViewedContent(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

// UpdateContent обновляет контент (только для преподавателя)
func (h *ContentHandler) UpdateContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	var req CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.contentService.UpdateContent(contentID, &services.CreateContentRequest{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Category:    req.Category,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Content updated successfully"})
}

// DeleteContent удаляет контент (только для преподавателя)
func (h *ContentHandler) DeleteContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	if err := h.contentService.DeleteContent(contentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Content deleted successfully"})
}

// GetContentCategories возвращает список доступных категорий
func (h *ContentHandler) GetContentCategories(c *gin.Context) {
	categories := h.contentService.GetContentCategories()
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetContentTypes возвращает список доступных типов контента
func (h *ContentHandler) GetContentTypes(c *gin.Context) {
	types := h.contentService.GetContentTypes()
	c.JSON(http.StatusOK, gin.H{"types": types})
}

// GetContentStats возвращает статистику контента (только для преподавателя)
func (h *ContentHandler) GetContentStats(c *gin.Context) {
	stats, err := h.contentService.GetContentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
