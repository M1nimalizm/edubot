package handlers

import (
	"net/http"
	"strings"

	"edubot/internal/models"
	"edubot/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware создает middleware для авторизации
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
        // Получаем токен из заголовка Authorization или cookie
        var token string
        if authHeader := c.GetHeader("Authorization"); authHeader != "" {
            parts := strings.Split(authHeader, " ")
            if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
                token = parts[1]
            }
        }
        if token == "" {
            if cookie, err := c.Cookie("jwt"); err == nil {
                token = cookie
            }
        }
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

		// Валидируем токен
		user, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте (строгие типы)
		c.Set("user", user)
		c.Set("user_id", user.ID) // uuid.UUID
		c.Set("telegram_id", user.TelegramID)
		c.Set("user_role", user.Role) // models.UserRole

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

		// Сохраняем данные пользователя в контексте (строгие типы)
		c.Set("user", user)
		c.Set("user_id", user.ID) // uuid.UUID
		c.Set("telegram_id", user.TelegramID)
		c.Set("user_role", user.Role) // models.UserRole

		c.Next()
	}
}

// TeacherOnlyMiddleware создает middleware только для преподавателей
func TeacherOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		role, ok := roleVal.(models.UserRole)
		if !exists || !ok || role != models.RoleTeacher {
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
		roleVal, exists := c.Get("user_role")
		role, ok := roleVal.(models.UserRole)
		if !exists || !ok || role != models.RoleStudent {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Student role required."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRoles разрешает доступ только указанным ролям
func RequireRoles(allowed ...models.UserRole) gin.HandlerFunc {
	allowedSet := make(map[models.UserRole]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		role, ok := roleVal.(models.UserRole)
		if !exists || !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}
		if _, ok := allowedSet[role]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORSMiddleware создает middleware для CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Разрешаем все домены для Telegram WebApp
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Заголовки для Telegram WebApp
		c.Header("X-Frame-Options", "ALLOWALL")
		c.Header("Content-Security-Policy", "frame-ancestors *")
		c.Header("Referrer-Policy", "no-referrer-when-downgrade")

		// Дополнительные заголовки для Telegram
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequireHTMLRoles редиректит на главную, если роль не подходит
func RequireHTMLRoles(allowed ...models.UserRole) gin.HandlerFunc {
	allowedSet := make(map[models.UserRole]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		role, ok := roleVal.(models.UserRole)
		if !exists || !ok {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		if _, ok := allowedSet[role]; !ok {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
