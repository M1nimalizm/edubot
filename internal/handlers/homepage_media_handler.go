package handlers

import (
	"edubot/internal/models"
	"edubot/internal/services"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HomepageMediaHandler struct {
	mediaService services.HomepageMediaService
}

func NewHomepageMediaHandler(mediaService services.HomepageMediaService) *HomepageMediaHandler {
	return &HomepageMediaHandler{
		mediaService: mediaService,
	}
}

// UploadMedia загружает медиафайл для главной страницы
func (h *HomepageMediaHandler) UploadMedia(c *gin.Context) {
	// Получаем тип медиафайла из параметра
	mediaTypeStr := c.Param("type")
	mediaType := models.HomepageMediaType(mediaTypeStr)
	
	// Проверяем валидность типа
	if mediaType != models.HomepageMediaTypeTeacherPhoto && 
	   mediaType != models.HomepageMediaTypeWelcomeVideo &&
	   mediaType != models.HomepageMediaTypeHeroImage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media type"})
		return
	}
	
	// Получаем файл из формы
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	
	// Проверяем размер файла (максимум 50MB)
	if file.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum size is 50MB"})
		return
	}
	
	// Проверяем тип файла
	ext := filepath.Ext(file.Filename)
	if mediaType == models.HomepageMediaTypeTeacherPhoto || mediaType == models.HomepageMediaTypeHeroImage {
		// Для изображений разрешены только определенные форматы
		allowedExts := []string{".jpg", ".jpeg", ".png", ".webp"}
		allowed := false
		for _, allowedExt := range allowedExts {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Allowed: jpg, jpeg, png, webp"})
			return
		}
	} else if mediaType == models.HomepageMediaTypeWelcomeVideo {
		// Для видео разрешены определенные форматы
		allowedExts := []string{".mp4", ".webm", ".ogg"}
		allowed := false
		for _, allowedExt := range allowedExts {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Allowed: mp4, webm, ogg"})
			return
		}
	}
	
	// Загружаем файл
	media, err := h.mediaService.UploadFile(file, mediaType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"media":   media,
	})
}

// GetMedia получает медиафайл по ID
func (h *HomepageMediaHandler) GetMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	
	media, err := h.mediaService.GetMediaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}
	
	c.JSON(http.StatusOK, media)
}

// GetActiveMedia получает активный медиафайл по типу
func (h *HomepageMediaHandler) GetActiveMedia(c *gin.Context) {
	mediaTypeStr := c.Param("type")
	mediaType := models.HomepageMediaType(mediaTypeStr)
	
	media, err := h.mediaService.GetActiveMedia(mediaType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active media found for this type"})
		return
	}
	
	c.JSON(http.StatusOK, media)
}

// ListMedia получает список всех медиафайлов
func (h *HomepageMediaHandler) ListMedia(c *gin.Context) {
	media, err := h.mediaService.ListMedia()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media list"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"media": media})
}

// SetActiveMedia устанавливает активный медиафайл
func (h *HomepageMediaHandler) SetActiveMedia(c *gin.Context) {
	mediaTypeStr := c.Param("type")
	mediaType := models.HomepageMediaType(mediaTypeStr)
	
	var req struct {
		MediaID string `json:"media_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	mediaID, err := uuid.Parse(req.MediaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	
	err = h.mediaService.SetActiveMedia(mediaType, mediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active media: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Active media updated successfully"})
}

// DeleteMedia удаляет медиафайл
func (h *HomepageMediaHandler) DeleteMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	
	err = h.mediaService.DeleteMedia(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete media: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Media deleted successfully"})
}

// ServeMedia отдает медиафайл по URL
func (h *HomepageMediaHandler) ServeMedia(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename required"})
		return
	}
	
	// Безопасность: проверяем, что файл не содержит путь
	if filepath.Base(filename) != filename {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}
	
	// Ищем медиафайл в базе данных
	mediaList, err := h.mediaService.ListMedia()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media list"})
		return
	}
	
	var targetMedia *models.HomepageMedia
	for _, media := range mediaList {
		if media.Filename == filename {
			targetMedia = media
			break
		}
	}
	
	if targetMedia == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}
	
	// Отдаем файл
	c.File(targetMedia.Path)
}
