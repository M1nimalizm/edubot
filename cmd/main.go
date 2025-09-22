package main

import (
	"fmt"
	"log"
	"net/http"

	"edubot/internal/config"
	"edubot/internal/handlers"
	"edubot/internal/repository"
	"edubot/internal/services"
	"edubot/pkg/database"
	"edubot/pkg/storage"
	"edubot/pkg/telegram"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к базе данных
	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Создаем пользователя-преподавателя по умолчанию
	if err := db.CreateDefaultTeacher(cfg.TeacherTelegramID); err != nil {
		log.Printf("Failed to create default teacher: %v", err)
	}

	// Инициализируем файловое хранилище
	_, err = storage.NewStorage(cfg.UploadPath, cfg.MaxFileSize, cfg.MaxUserStorage)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Инициализируем Telegram бота
	telegramBot, err := telegram.NewBot(cfg.TelegramBotToken, cfg.TelegramWebhookURL)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Устанавливаем команды бота
	if err := telegramBot.SetCommands(); err != nil {
		log.Printf("Failed to set bot commands: %v", err)
	}

	// Устанавливаем webhook если указан URL
	if cfg.TelegramWebhookURL != "" {
		if err := telegramBot.SetWebhook(); err != nil {
			log.Printf("Failed to set webhook: %v", err)
		}
	}

	// Создаем репозитории
	userRepo := repository.NewUserRepository(db.DB)
	trialRepo := repository.NewTrialRequestRepository(db.DB)
	assignmentRepo := repository.NewAssignmentRepository(db.DB)

	// Создаем сервисы
	authService := services.NewAuthService(userRepo, trialRepo, telegramBot, cfg.JWTSecret, cfg.TeacherTelegramID)
	assignmentService := services.NewAssignmentService(assignmentRepo, userRepo, telegramBot)

	// Создаем обработчики
	authHandler := handlers.NewAuthHandler(authService)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentService)

	// Настраиваем Gin
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware
	router.Use(handlers.CORSMiddleware())

	// Статические файлы
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Специальный endpoint для Telegram WebApp
	router.GET("/telegram-check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":       "telegram_ready",
			"webapp_url":   "https://edubot-0g05.onrender.com",
			"bot_username": "EduBot_by_Pugachev_bot",
		})
	})

	// Главная страница
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "EduBot - Образовательная платформа",
		})
	})

	// Панель управления учителя
	router.GET("/teacher-dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-dashboard.html", gin.H{
			"title": "Панель управления - EduBot",
		})
	})

	// Страницы для учителя
	router.GET("/teacher/assignments/create", func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-assignments.html", gin.H{
			"title": "Создание задания - EduBot",
		})
	})

	router.GET("/teacher/content/create", func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-content.html", gin.H{
			"title": "Добавление материалов - EduBot",
		})
	})

	router.GET("/teacher/students", func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-students.html", gin.H{
			"title": "Ученики - EduBot",
		})
	})

	// Страницы для учеников
	router.GET("/student-dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "student-dashboard.html", gin.H{
			"title": "Мои задания - EduBot",
		})
	})

	router.GET("/student-progress", func(c *gin.Context) {
		c.HTML(http.StatusOK, "student-progress.html", gin.H{
			"title": "Мой прогресс - EduBot",
		})
	})

	// API маршруты
	api := router.Group("/api")

	// Публичные маршруты (доступны гостям)
	public := api.Group("/public")
	{
		public.POST("/auth/telegram", authHandler.TelegramAuth)
		public.POST("/trial-request", handlers.GuestMiddleware(authService), authHandler.SubmitTrialRequest)
		public.POST("/register-student", authHandler.RegisterStudentByCode)
	}

	// Публичные маршруты для панели управления учителя (без авторизации для простоты)
	teacherPublic := api.Group("/teacher")
	{
		teacherPublic.GET("/trial-requests", authHandler.GetTrialRequests)
		teacherPublic.GET("/stats", authHandler.GetStats)
		teacherPublic.POST("/invite-code", authHandler.GenerateInviteCode)
	}

	// Защищенные маршруты (требуют авторизации)
	protected := api.Group("/")
	protected.Use(handlers.AuthMiddleware(authService))
	{
		// Профиль пользователя
		protected.GET("/profile", authHandler.GetProfile)
		protected.POST("/register-student", authHandler.RegisterStudent)

		// Задания для учеников
		protected.GET("/assignments", assignmentHandler.GetStudentAssignments)
		protected.GET("/assignments/:id", assignmentHandler.GetAssignment)
		protected.POST("/assignments/:id/complete", assignmentHandler.MarkAssignmentCompleted)
		protected.GET("/assignments/upcoming", assignmentHandler.GetUpcomingDeadlines)

		// Комментарии к заданиям
		protected.GET("/assignments/:id/comments", assignmentHandler.GetComments)
		protected.POST("/assignments/:id/comments", assignmentHandler.AddComment)

		// Контент
		protected.GET("/content/:id", assignmentHandler.GetContent)
		protected.GET("/content/subject/:subject/grade/:grade", assignmentHandler.GetContentBySubject)

		// Прогресс ученика
		protected.GET("/progress", assignmentHandler.GetStudentProgress)
	}

	// Маршруты только для преподавателей (защищенные)
	teacher := api.Group("/teacher")
	teacher.Use(handlers.AuthMiddleware(authService))
	teacher.Use(handlers.TeacherOnlyMiddleware())
	{
		// Генерация кодов приглашения (дублирующий маршрут убран)
		teacher.POST("/invite-codes", authHandler.GenerateInviteCode)

		// Управление учениками
		teacher.GET("/students", authHandler.GetStudents)

		// Управление заданиями
		teacher.POST("/assignments", assignmentHandler.CreateAssignment)
		teacher.GET("/assignments", assignmentHandler.GetTeacherAssignments)
		teacher.PUT("/assignments/:id", assignmentHandler.UpdateAssignment)
		teacher.DELETE("/assignments/:id", assignmentHandler.DeleteAssignment)

		// Управление контентом
		teacher.POST("/content", assignmentHandler.CreateContent)
		teacher.GET("/content", assignmentHandler.GetTeacherContent)
		teacher.PUT("/content/:id", assignmentHandler.UpdateContent)
		teacher.DELETE("/content/:id", assignmentHandler.DeleteContent)
	}

	// Webhook для Telegram
	router.GET("/webhook", func(c *gin.Context) {
		// Telegram проверяет доступность webhook
		c.JSON(http.StatusOK, gin.H{"status": "webhook_ready"})
	})

	router.POST("/webhook", func(c *gin.Context) {
		var update map[string]interface{}
		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Обрабатываем обновление от Telegram
		if telegramBot != nil {
			telegramBot.ProcessUpdate(update)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Запускаем сервер
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Starting EduBot server on %s", addr)
	log.Printf("Teacher Telegram ID: %d", cfg.TeacherTelegramID)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
