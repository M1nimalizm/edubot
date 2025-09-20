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
	fileStorage, err := storage.NewStorage(cfg.UploadPath, cfg.MaxFileSize, cfg.MaxUserStorage)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Инициализируем Telegram бота
	telegramBot, err := telegram.NewBot(cfg.TelegramBotToken, cfg.TelegramWebhookURL)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
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
	submissionRepo := repository.NewSubmissionRepository(db.DB)
	contentRepo := repository.NewContentRepository(db.DB)
	attachmentRepo := repository.NewAttachmentRepository(db.DB)

	// Создаем сервисы
	authService := services.NewAuthService(userRepo, trialRepo, telegramBot, cfg.JWTSecret, cfg.TeacherTelegramID)
	assignmentService := services.NewAssignmentService(assignmentRepo, submissionRepo, attachmentRepo, userRepo, fileStorage, telegramBot)
	contentService := services.NewContentService(contentRepo, attachmentRepo, fileStorage)

	// Создаем обработчики
	authHandler := handlers.NewAuthHandler(authService)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentService)
	contentHandler := handlers.NewContentHandler(contentService)

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

	// Главная страница
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "EduBot - Образовательная платформа",
		})
	})

	// API маршруты
	api := router.Group("/api")

	// Публичные маршруты (доступны гостям)
	public := api.Group("/public")
	{
		public.POST("/auth/telegram", authHandler.TelegramAuth)
		public.POST("/trial-request", handlers.GuestMiddleware(authService), authHandler.SubmitTrialRequest)
		public.GET("/content", contentHandler.ListContent)
		public.GET("/content/categories", contentHandler.GetContentCategories)
		public.GET("/content/types", contentHandler.GetContentTypes)
		public.GET("/content/search", contentHandler.SearchContent)
		public.GET("/content/:id", contentHandler.GetContent)
	}

	// Защищенные маршруты (требуют авторизации)
	protected := api.Group("/")
	protected.Use(handlers.AuthMiddleware(authService))
	{
		// Профиль пользователя
		protected.GET("/profile", authHandler.GetProfile)
		protected.POST("/register-student", authHandler.RegisterStudent)

		// Задания
		protected.GET("/assignments", assignmentHandler.GetAssignments)
		protected.GET("/assignments/:id", assignmentHandler.GetAssignment)
		protected.POST("/assignments/:id/submit", assignmentHandler.SubmitSolution)
		protected.GET("/assignments/upcoming", assignmentHandler.GetUpcomingDeadlines)

		// Просмотренный контент
		protected.GET("/content/viewed", contentHandler.GetViewedContent)
	}

	// Маршруты только для преподавателей
	teacher := api.Group("/teacher")
	teacher.Use(handlers.AuthMiddleware(authService))
	teacher.Use(handlers.TeacherOnlyMiddleware())
	{
		// Управление заявками на пробные занятия
		teacher.GET("/trial-requests", func(c *gin.Context) {
			// TODO: Реализовать получение заявок
			c.JSON(http.StatusOK, gin.H{"message": "Trial requests endpoint"})
		})

		// Генерация кодов приглашения
		teacher.POST("/invite-codes", authHandler.GenerateInviteCode)

		// Управление заданиями
		teacher.POST("/assignments", assignmentHandler.CreateAssignment)
		teacher.GET("/assignments/pending", assignmentHandler.GetPendingSubmissions)
		teacher.POST("/submissions/:id/grade", assignmentHandler.GradeSubmission)
		teacher.POST("/assignments/reminders", assignmentHandler.SendDeadlineReminders)

		// Управление контентом
		teacher.POST("/content", contentHandler.CreateContent)
		teacher.PUT("/content/:id", contentHandler.UpdateContent)
		teacher.DELETE("/content/:id", contentHandler.DeleteContent)
		teacher.GET("/content/stats", contentHandler.GetContentStats)
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
