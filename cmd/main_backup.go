package main

import (
	"fmt"
	"log"
	"net/http"

	"edubot/internal/config"
	"edubot/internal/handlers"
	"edubot/internal/models"
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

	log.Printf("Configuration loaded successfully")
	log.Printf("Port: %s", cfg.Port)
	log.Printf("Host: %s", cfg.Host)
	log.Printf("Base URL: %s", cfg.BaseURL)

	// Подключаемся к базе данных
	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Printf("Database connected successfully")

	// Создаем пользователя-преподавателя по умолчанию
	if err := db.CreateDefaultTeacher(cfg.TeacherTelegramID); err != nil {
		log.Printf("Failed to create default teacher: %v", err)
	}
	log.Printf("Default teacher setup completed")

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

	// Устанавливаем команды бота (после инициализации сервисов подключим колбэки)
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
	groupRepo := repository.NewGroupRepository(db.DB)
	mediaRepo := repository.NewMediaRepository(db.DB)
	homepageMediaRepo := repository.NewHomepageMediaRepository(db.DB)

	// Создаем сервисы
	authService := services.NewAuthService(
		userRepo,
		trialRepo,
		telegramBot,
		cfg.JWTSecret,
		cfg.TeacherTelegramID,
		cfg.TeacherTelegramIDs,
		cfg.TeacherPassword,
	)
	mediaService := services.NewMediaService(mediaRepo, userRepo, telegramBot, assignmentRepo)
	assignmentService := services.NewAssignmentService(assignmentRepo, userRepo, mediaService, telegramBot)
	groupService := services.NewGroupService(groupRepo, userRepo, assignmentRepo, telegramBot)
	homepageMediaService := services.NewHomepageMediaService(homepageMediaRepo, cfg.BaseURL, "./uploads/homepage")

	// Создаем обработчики
	authHandler := handlers.NewAuthHandler(authService)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentService)
	groupHandler := handlers.NewGroupHandler(groupService)
	mediaHandler := handlers.NewMediaHandler(mediaService)
	homepageMediaHandler := handlers.NewHomepageMediaHandler(homepageMediaService)

	// Подключаем колбэки бота к бэкенду (после инициализации сервисов)
	telegramBot.SetAssignStudent(func(teacherTelegramID int64, telegramID *int64, username string, grade *int, subjects string) error {
		teacher, err := userRepo.GetByTelegramID(teacherTelegramID)
		if err != nil {
			return fmt.Errorf("teacher not found")
		}
		_, err = authService.AssignStudentToTeacher(teacher.ID, services.AssignStudentParams{
			TelegramID: telegramID,
			Username:   username,
			Grade:      grade,
			Subjects:   subjects,
		})
		return err
	})
	telegramBot.SetGetUserRole(func(telegramID int64) string {
		u, err := userRepo.GetByTelegramID(telegramID)
		if err != nil || u == nil {
			return "guest"
		}
		return string(u.Role)
	})

	telegramBot.SetListTeacherGroups(func(teacherTelegramID int64) ([]struct {
		ID   string
		Name string
	}, error) {
		teacher, err := userRepo.GetByTelegramID(teacherTelegramID)
		if err != nil {
			return nil, err
		}
		gs, err := groupService.ListGroups(teacher.ID)
		if err != nil {
			return nil, err
		}
		res := make([]struct {
			ID   string
			Name string
		}, 0, len(gs))
		for _, g := range gs {
			res = append(res, struct {
				ID   string
				Name string
			}{ID: g.ID.String(), Name: g.Name})
		}
		return res, nil
	})

	// Настраиваем Gin
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware
	router.Use(handlers.CORSMiddleware())

	// Статические файлы и HTML шаблоны
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Публичные маршруты для медиафайлов главной страницы
	router.GET("/media/homepage/:filename", homepageMediaHandler.ServeMedia)
	router.GET("/api/public/homepage-media/:type", homepageMediaHandler.GetActiveMedia)

	// Специальный endpoint для Telegram WebApp
	router.GET("/telegram-check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":       "telegram_ready",
            "webapp_url":   "https://edubot-0g05.onrender.com/app",
			"bot_username": "EduBot_by_Pugachev_bot",
		})
	})

	// Тестовый маршрут для проверки работы сервера
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"message":  "EduBot server is running",
			"base_url": cfg.BaseURL,
			"port":     cfg.Port,
		})
	})

	// Главная страница
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "EduBot - Образовательная платформа",
		})
	})

	// Панель управления учителя (HTML guard)
	router.GET("/teacher-dashboard", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleTeacher), func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-dashboard.html", gin.H{
			"title": "Панель управления - EduBot",
		})
	})

	// Страницы для учителя
	router.GET("/teacher/assignments/create", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleTeacher), func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-assignments.html", gin.H{
			"title": "Создание задания - EduBot",
		})
	})

	router.GET("/teacher-submissions", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleTeacher), func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-submissions.html", gin.H{
			"title": "Проверка заданий - EduBot",
		})
	})

	router.GET("/teacher/content/create", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleTeacher), func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-content.html", gin.H{
			"title": "Добавление материалов - EduBot",
		})
	})

	router.GET("/teacher/students", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleTeacher), func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-students.html", gin.H{
			"title": "Ученики - EduBot",
		})
	})

	router.GET("/teacher-groups", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleTeacher), func(c *gin.Context) {
		c.HTML(http.StatusOK, "teacher-groups.html", gin.H{
			"title": "Группы - EduBot",
		})
	})

	router.GET("/homepage-media", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleTeacher), func(c *gin.Context) {
		c.HTML(http.StatusOK, "homepage-media.html", gin.H{
			"title": "Управление медиафайлами - EduBot",
		})
	})

	// Страницы для учеников
	router.GET("/student-dashboard", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleStudent), func(c *gin.Context) {
		c.HTML(http.StatusOK, "student-dashboard.html", gin.H{
			"title": "Мои задания - EduBot",
		})
	})

	router.GET("/student-progress", handlers.AuthMiddleware(authService), handlers.RequireHTMLRoles(models.RoleStudent), func(c *gin.Context) {
		c.HTML(http.StatusOK, "student-progress.html", gin.H{
			"title": "Мой прогресс - EduBot",
		})
	})

	// API маршруты
	api := router.Group("/api")

	// Публичные маршруты (доступны гостям)
	// Публичные маршруты: временно отключены авторизация и регистрация
	public := api.Group("/public")
	{
		// Публичные медиафайлы (приветственные ролики и т.д.)
		public.GET("/media", mediaHandler.GetPublicMedia)
		// Авторизация через Telegram WebApp/Desktop
		public.POST("/auth/telegram", authHandler.TelegramAuth)
		// Заявка на пробное занятие
		public.POST("/trial-request", authHandler.SubmitTrialRequest)
	}

	// Совместимость: /api/media/public (чтобы не перехватывалось /media/:id)
	api.GET("/media/public", mediaHandler.GetPublicMedia)
	_ = public

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

		// Задания для учеников (student only)
		protected.GET("/assignments", handlers.StudentOnlyMiddleware(), assignmentHandler.GetStudentAssignments)
		protected.GET("/assignments/:id", handlers.StudentOnlyMiddleware(), assignmentHandler.GetAssignment)
		protected.POST("/assignments/:id/complete", handlers.StudentOnlyMiddleware(), assignmentHandler.MarkAssignmentCompleted)
		protected.GET("/assignments/upcoming", handlers.StudentOnlyMiddleware(), assignmentHandler.GetUpcomingDeadlines)

		// Комментарии к заданиям
		protected.GET("/assignments/:id/comments", assignmentHandler.GetComments)
		protected.POST("/assignments/:id/comments", assignmentHandler.AddComment)

		// Медиафайлы заданий
		protected.POST("/assignments/:id/media", assignmentHandler.AddAssignmentMedia)
		protected.GET("/assignments/:id/media", assignmentHandler.GetAssignmentMedia)

		// Сдача заданий
		protected.POST("/assignments/:id/submit", handlers.StudentOnlyMiddleware(), assignmentHandler.SubmitAssignment)

		// Медиафайлы submissions
		protected.GET("/submissions/:id/media", handlers.RequireRoles(models.RoleStudent, models.RoleTeacher), assignmentHandler.GetSubmissionMedia)

		// Фидбэк учителя
		protected.POST("/submissions/:id/feedback", handlers.TeacherOnlyMiddleware(), assignmentHandler.AddFeedbackMedia)
		protected.GET("/submissions/:id/feedback", handlers.RequireRoles(models.RoleStudent, models.RoleTeacher), assignmentHandler.GetFeedbackMedia)

		// Submissions для учителя
		protected.GET("/teacher/submissions", handlers.TeacherOnlyMiddleware(), assignmentHandler.GetTeacherSubmissions)
		protected.POST("/teacher/submissions/:id/feedback", handlers.TeacherOnlyMiddleware(), assignmentHandler.SubmitTeacherFeedback)

		// Контент
		protected.GET("/content/:id", assignmentHandler.GetContent)
		protected.GET("/content/subject/:subject/grade/:grade", assignmentHandler.GetContentBySubject)

		// Прогресс ученика
		protected.GET("/progress", assignmentHandler.GetStudentProgress)

		// Медиафайлы
		protected.POST("/media", mediaHandler.CreateMedia)
		protected.POST("/media/upload", mediaHandler.UploadMedia)
		protected.GET("/media/:id", mediaHandler.GetMedia)
		protected.GET("/media/:id/stream", mediaHandler.StreamMedia)
		protected.GET("/media/:id/thumbnail", mediaHandler.GetThumbnail)
		protected.GET("/media", mediaHandler.GetUserMedia)
		protected.GET("/media/entity/:entity_type/:entity_id", mediaHandler.GetEntityMedia)
		protected.PUT("/media/:id", mediaHandler.UpdateMedia)
		protected.DELETE("/media/:id", mediaHandler.DeleteMedia)
		protected.GET("/media/:id/views", mediaHandler.GetMediaViews)
		protected.POST("/media/:id/access", mediaHandler.GrantAccess)
		protected.DELETE("/media/:id/access/:user_id", mediaHandler.RevokeAccess)
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
		teacher.POST("/students", authHandler.CreateStudentByTeacher)
		teacher.POST("/students/assign", authHandler.AssignStudentToTeacher)

		// Группы
		teacher.POST("/groups", groupHandler.CreateGroup)
		teacher.GET("/groups", groupHandler.ListGroups)
		teacher.POST("/groups/:id/members", groupHandler.AddMember)
		teacher.DELETE("/groups/:id/members/:user_id", groupHandler.RemoveMember)
		teacher.POST("/groups/:id/assignments", groupHandler.AssignHomework)

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

		// Управление медиафайлами главной страницы
		teacher.POST("/homepage-media/:type", homepageMediaHandler.UploadMedia)
		teacher.GET("/homepage-media", homepageMediaHandler.ListMedia)
		teacher.GET("/homepage-media/:id", homepageMediaHandler.GetMedia)
		teacher.PUT("/homepage-media/:type/active", homepageMediaHandler.SetActiveMedia)
		teacher.DELETE("/homepage-media/:id", homepageMediaHandler.DeleteMedia)
	}

	// Выбор роли после Telegram-авторизации (без пароля)
	// Отключено: выбор роли/учительская авторизация
	// api.POST("/auth/select-role", handlers.AuthMiddleware(authService), authHandler.SelectRole)

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
	log.Printf("Base URL: %s", cfg.BaseURL)
	log.Printf("Database path: %s", cfg.DBPath)
	log.Printf("Upload path: %s", cfg.UploadPath)
	log.Printf("Teacher Telegram ID: %d", cfg.TeacherTelegramID)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
