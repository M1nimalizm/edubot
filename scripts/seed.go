package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"edubot/internal/models"
)

func main() {
	// Подключаемся к базе данных
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автомиграция
	err = db.AutoMigrate(
		&models.User{},
		&models.Assignment{},
		&models.AssignmentTarget{},
		&models.Submission{},
		&models.Feedback{},
		&models.ChatThread{},
		&models.Message{},
		&models.Notification{},
		&models.Group{},
		&models.GroupMember{},
		&models.Media{},
		&models.Draft{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Создаем тестовых пользователей
	teacherID := uuid.New()
	student1ID := uuid.New()
	student2ID := uuid.New()
	student3ID := uuid.New()

	users := []models.User{
		{
			ID:         teacherID,
			TelegramID: 123456789,
			Username:   "teacher_pugach",
			FirstName:  "Александр",
			LastName:   "Пугачев",
			Role:       models.RoleTeacher,
			Grade:      0,
			Subjects:   "physics,math",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         student1ID,
			TelegramID: 987654321,
			Username:   "student_ivan",
			FirstName:  "Иван",
			LastName:   "Иванов",
			Role:       models.RoleStudent,
			Grade:      11,
			Subjects:   "physics,math",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         student2ID,
			TelegramID: 111222333,
			Username:   "student_maria",
			FirstName:  "Мария",
			LastName:   "Петрова",
			Role:       models.RoleStudent,
			Grade:      10,
			Subjects:   "physics",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         student3ID,
			TelegramID: 444555666,
			Username:   "student_alex",
			FirstName:  "Алексей",
			LastName:   "Сидоров",
			Role:       models.RoleStudent,
			Grade:      11,
			Subjects:   "math",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user %s: %v", user.Username, err)
		}
	}

	// Создаем группу
	groupID := uuid.New()
	group := models.Group{
		ID:        groupID,
		Name:      "11 класс - Физика",
		Subject:   "physics",
		Grade:     11,
		Level:     3,
		TeacherID: teacherID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&group).Error; err != nil {
		log.Printf("Failed to create group: %v", err)
	}

	// Добавляем студентов в группу
	groupMembers := []models.GroupMember{
		{
			ID:       uuid.New(),
			GroupID:  groupID,
			UserID:   student1ID,
			Role:     "student",
			JoinedAt: time.Now(),
		},
		{
			ID:       uuid.New(),
			GroupID:  groupID,
			UserID:   student2ID,
			Role:     "student",
			JoinedAt: time.Now(),
		},
	}

	for _, member := range groupMembers {
		if err := db.Create(&member).Error; err != nil {
			log.Printf("Failed to create group member: %v", err)
		}
	}

	// Создаем задания
	assignments := []models.Assignment{
		{
			ID:          uuid.New(),
			Title:       "Кинематика - Равномерное движение",
			Description: "Решить задачи на равномерное прямолинейное движение. Внимательно изучите формулы и единицы измерения.",
			Subject:     "physics",
			Grade:       11,
			Level:       2,
			TeacherID:   teacherID,
			GroupID:     &groupID,
			DueDate:     time.Now().Add(7 * 24 * time.Hour),
			Status:      "active",
			CreatedBy:   teacherID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Динамика - Законы Ньютона",
			Description: "Применить законы Ньютона для решения задач на движение тел под действием сил.",
			Subject:     "physics",
			Grade:       11,
			Level:       3,
			TeacherID:   teacherID,
			GroupID:     &groupID,
			DueDate:     time.Now().Add(14 * 24 * time.Hour),
			Status:      "active",
			CreatedBy:   teacherID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Индивидуальное задание - Математика",
			Description: "Решить систему уравнений и построить график функции.",
			Subject:     "math",
			Grade:       11,
			Level:       4,
			TeacherID:   teacherID,
			StudentID:   &student3ID,
			DueDate:     time.Now().Add(5 * 24 * time.Hour),
			Status:      "active",
			CreatedBy:   teacherID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, assignment := range assignments {
		if err := db.Create(&assignment).Error; err != nil {
			log.Printf("Failed to create assignment: %v", err)
		}
	}

	// Создаем AssignmentTarget для групповых заданий
	for _, assignment := range assignments[:2] { // Первые два задания - групповые
		for _, member := range groupMembers {
			target := models.AssignmentTarget{
				ID:           uuid.New(),
				AssignmentID: assignment.ID,
				StudentID:    member.UserID,
				Status:       models.AssignmentTargetStatusPending,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if err := db.Create(&target).Error; err != nil {
				log.Printf("Failed to create assignment target: %v", err)
			}
		}
	}

	// Создаем AssignmentTarget для индивидуального задания
	individualTarget := models.AssignmentTarget{
		ID:           uuid.New(),
		AssignmentID: assignments[2].ID,
		StudentID:    student3ID,
		Status:       models.AssignmentTargetStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&individualTarget).Error; err != nil {
		log.Printf("Failed to create individual assignment target: %v", err)
	}

	// Создаем чат между студентом и преподавателем
	chatThreadID := uuid.New()
	chatThread := models.ChatThread{
		ID:           chatThreadID,
		Type:         models.ChatThreadTypeStudentTeacher,
		StudentID:    &student1ID,
		TeacherID:    teacherID,
		LastMessageAt: &[]time.Time{time.Now()}[0],
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&chatThread).Error; err != nil {
		log.Printf("Failed to create chat thread: %v", err)
	}

	// Создаем несколько сообщений в чате
	messages := []models.Message{
		{
			ID:        uuid.New(),
			ThreadID:  chatThreadID,
			AuthorID:  student1ID,
			Text:      &[]string{"Здравствуйте! У меня вопрос по задаче на кинематику."}[0],
			Kind:      models.MessageKindMessage,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			ThreadID:  chatThreadID,
			AuthorID:  teacherID,
			Text:      &[]string{"Здравствуйте, Иван! Какой именно вопрос у вас возник?"}[0],
			Kind:      models.MessageKindMessage,
			CreatedAt: time.Now().Add(5 * time.Minute),
		},
		{
			ID:        uuid.New(),
			ThreadID:  chatThreadID,
			AuthorID:  student1ID,
			Text:      &[]string{"Не понимаю, как найти время движения, если известны только скорость и путь."}[0],
			Kind:      models.MessageKindMessage,
			CreatedAt: time.Now().Add(10 * time.Minute),
		},
	}

	for _, message := range messages {
		if err := db.Create(&message).Error; err != nil {
			log.Printf("Failed to create message: %v", err)
		}
	}

	// Создаем уведомления
	notifications := []models.Notification{
		{
			ID:        uuid.New(),
			UserID:    student1ID,
			Type:      models.NotificationTypeNewAssignment,
			Title:     "Новое задание",
			Message:   "Вам назначено новое задание: Кинематика - Равномерное движение",
			Payload:   fmt.Sprintf(`{"assignment_id":"%s"}`, assignments[0].ID.String()),
			Channel:   models.NotificationChannelBot,
			Status:    models.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			UserID:    student2ID,
			Type:      models.NotificationTypeNewAssignment,
			Title:     "Новое задание",
			Message:   "Вам назначено новое задание: Кинематика - Равномерное движение",
			Payload:   fmt.Sprintf(`{"assignment_id":"%s"}`, assignments[0].ID.String()),
			Channel:   models.NotificationChannelBot,
			Status:    models.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			UserID:    teacherID,
			Type:      models.NotificationTypeNewMessage,
			Title:     "Новое сообщение",
			Message:   "Иван Иванов написал сообщение в чате",
			Payload:   fmt.Sprintf(`{"thread_id":"%s","message_id":"%s"}`, chatThreadID.String(), messages[0].ID.String()),
			Channel:   models.NotificationChannelBot,
			Status:    models.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, notification := range notifications {
		if err := db.Create(&notification).Error; err != nil {
			log.Printf("Failed to create notification: %v", err)
		}
	}

	fmt.Println("✅ Seed data created successfully!")
	fmt.Printf("👨‍🏫 Teacher: %s (ID: %s)\n", users[0].FirstName+" "+users[0].LastName, teacherID.String())
	fmt.Printf("👨‍🎓 Student 1: %s (ID: %s)\n", users[1].FirstName+" "+users[1].LastName, student1ID.String())
	fmt.Printf("👩‍🎓 Student 2: %s (ID: %s)\n", users[2].FirstName+" "+users[2].LastName, student2ID.String())
	fmt.Printf("👨‍🎓 Student 3: %s (ID: %s)\n", users[3].FirstName+" "+users[3].LastName, student3ID.String())
	fmt.Printf("👥 Group: %s (ID: %s)\n", group.Name, groupID.String())
	fmt.Printf("📝 Assignments: %d created\n", len(assignments))
	fmt.Printf("💬 Chat messages: %d created\n", len(messages))
	fmt.Printf("🔔 Notifications: %d created\n", len(notifications))
}
