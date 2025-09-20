package handlers

import (
	"net/http"
	"strings"

	"edubot/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware создает middleware для авторизации
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Валидируем токен
		user, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user", user)
		c.Set("user_id", user.ID.String())
		c.Set("telegram_id", user.TelegramID)
		c.Set("user_role", string(user.Role))

		c.Next()
	}
}

// GuestMiddleware создает middleware для гостевого доступа
func GuestMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Если токена нет, разрешаем доступ как гостю
			c.Next()
			return
		}

		// Проверяем формат "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]

		// Валидируем токен
		user, err := authService.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user", user)
		c.Set("user_id", user.ID.String())
		c.Set("telegram_id", user.TelegramID)
		c.Set("user_role", string(user.Role))

		c.Next()
	}
}

// TeacherOnlyMiddleware создает middleware только для преподавателей
func TeacherOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Teacher role required."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// StudentOnlyMiddleware создает middleware только для учеников
func StudentOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole != "student" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Student role required."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORSMiddleware создает middleware для CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Заголовки для Telegram WebApp
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Content-Security-Policy", "frame-ancestors 'self' https://web.telegram.org https://telegram.org")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
