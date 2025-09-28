package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Получаем порт из переменной окружения (Render требует это)
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	// Настраиваем Gin для продакшена
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()

	// Простой тестовый маршрут
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "EduBot server is running!",
			"status": "ok",
			"port": port,
		})
	})

	// Тестовый маршрут для проверки работы сервера
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "EduBot server is running",
			"port": port,
		})
	})

	// Запускаем сервер
	log.Printf("Starting EduBot server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
