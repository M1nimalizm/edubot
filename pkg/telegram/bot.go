package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет Telegram бота
type Bot struct {
	api           *tgbotapi.BotAPI
	token         string
	webhook       string
	assignStudent func(teacherTelegramID int64, telegramID *int64, username string, grade *int, subjects string) error
	getUserRole   func(telegramID int64) string
	listGroups    func(teacherTelegramID int64) ([]struct {
		ID   string
		Name string
	}, error)
}

// NewBot создает новый экземпляр бота
func NewBot(token, webhook string) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = false // Включаем в режиме разработки

	return &Bot{
		api:     bot,
		token:   token,
		webhook: webhook,
	}, nil
}

// SetAssignStudent callback to backend
func (b *Bot) SetAssignStudent(cb func(teacherTelegramID int64, telegramID *int64, username string, grade *int, subjects string) error) {
	b.assignStudent = cb
}

// SetGetUserRole callback
func (b *Bot) SetGetUserRole(cb func(telegramID int64) string) { b.getUserRole = cb }

// SetListTeacherGroups callback
func (b *Bot) SetListTeacherGroups(cb func(teacherTelegramID int64) ([]struct {
	ID   string
	Name string
}, error)) {
	b.listGroups = cb
}

// SetWebhook устанавливает webhook для бота
func (b *Bot) SetWebhook() error {
	webhookConfig, err := tgbotapi.NewWebhook(b.webhook)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}
	_, err = b.api.Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}
	return nil
}

// SetCommands устанавливает команды бота
func (b *Bot) SetCommands() error {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "🚀 Начать работу с ботом",
		},
		{
			Command:     "help",
			Description: "ℹ️ Получить помощь по использованию",
		},
		{
			Command:     "app",
			Description: "📱 Открыть приложение EduBot",
		},
		{
			Command:     "info",
			Description: "👨‍🏫 Информация о преподавателе",
		},
	}

	setCommands := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.api.Request(setCommands)
	if err != nil {
		return fmt.Errorf("failed to set commands: %w", err)
	}
	return nil
}

// SendWelcomeToNewUser отправляет приветственное сообщение новому пользователю
func (b *Bot) SendWelcomeToNewUser(chatID int64, firstName string) error {
	text := fmt.Sprintf(`👋 Привет, %s! Добро пожаловать в EduBot!

🎓 Меня зовут Саша, я преподаватель физики и математики с 5-летним опытом подготовки к ЕГЭ.

📚 В моем приложении ты можешь:
• Узнать обо мне и моих методах обучения
• Записаться на пробное занятие
• Получить доступ к образовательным материалам
• Отслеживать свой прогресс

🚀 Начнем путь к успешной сдаче ЕГЭ вместе!

💡 <b>Быстрая навигация:</b>
• Используй /start для возврата в главное меню
• Используй /help для получения помощи
• Нажми кнопку ниже для перехода в приложение`, firstName)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Создаем клавиатуру с кнопкой "Открыть приложение"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🚀 Открыть приложение", "https://edubot-0g05.onrender.com"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Помощь", "help"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send welcome message: %w", err)
	}
	return nil
}

// SendMessage отправляет сообщение пользователю
func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// SendNotification отправляет уведомление преподавателю о новой заявке
func (b *Bot) SendTrialRequestNotification(teacherID int64, requestData map[string]interface{}) error {
	contactType := requestData["contact_type"].(string)
	contactValue := requestData["contact_value"].(string)

	var contactIcon, contactLabel string
	if contactType == "phone" {
		contactIcon = "📱"
		contactLabel = "Телефон"
	} else {
		contactIcon = "📲"
		contactLabel = "Telegram"
	}

	text := fmt.Sprintf(`
🎓 <b>Новая заявка на пробное занятие!</b>

👤 <b>Имя:</b> %s
📚 <b>Класс:</b> %d
📖 <b>Предмет:</b> %s
⭐ <b>Уровень:</b> %d/5
%s <b>%s:</b> %s

💬 <b>Комментарий:</b>
%s

🕐 <b>Время подачи:</b> %s
	`,
		requestData["name"],
		requestData["grade"],
		requestData["subject"],
		requestData["level"],
		contactIcon,
		contactLabel,
		contactValue,
		requestData["comment"],
		requestData["created_at"],
	)

	return b.SendMessage(teacherID, text)
}

// SendAssignmentNotification отправляет уведомление о новом задании
func (b *Bot) SendAssignmentNotification(chatID int64, assignmentTitle, subject string, deadline string) error {
	text := fmt.Sprintf(`
📝 <b>Новое задание!</b>

📖 <b>Предмет:</b> %s
📋 <b>Название:</b> %s
⏰ <b>Дедлайн:</b> %s

Переходите в приложение для просмотра деталей!
	`, subject, assignmentTitle, deadline)

	return b.SendMessage(chatID, text)
}

// SendDeadlineReminder отправляет напоминание о приближающемся дедлайне
func (b *Bot) SendDeadlineReminder(chatID int64, assignmentTitle string, hoursLeft int) error {
	text := fmt.Sprintf(`
⏰ <b>Напоминание о дедлайне!</b>

📋 <b>Задание:</b> %s
⏳ <b>Осталось:</b> %d часов

Не забудьте сдать задание вовремя!
	`, assignmentTitle, hoursLeft)

	return b.SendMessage(chatID, text)
}

// SendGradeNotification отправляет уведомление о проверенной работе
func (b *Bot) SendGradeNotification(chatID int64, assignmentTitle string, grade int, comments string) error {
	text := fmt.Sprintf(`
✅ <b>Работа проверена!</b>

📋 <b>Задание:</b> %s
⭐ <b>Оценка:</b> %d/5

💬 <b>Комментарии преподавателя:</b>
%s
	`, assignmentTitle, grade, comments)

	return b.SendMessage(chatID, text)
}

// GetUpdates получает обновления от Telegram
func (b *Bot) GetUpdates() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	return updates, nil
}

// ProcessUpdate обрабатывает входящее обновление
func (b *Bot) ProcessUpdate(update map[string]interface{}) {
	// Проверяем, есть ли callback query (нажатие на inline-кнопку)
	if callbackQuery, ok := update["callback_query"].(map[string]interface{}); ok {
		b.processCallbackQuery(callbackQuery)
		return
	}

	// Обрабатываем обычные сообщения
	message, ok := update["message"].(map[string]interface{})
	if !ok {
		return
	}

	text, _ := message["text"].(string)
	from, _ := message["from"].(map[string]interface{})
	chat, _ := message["chat"].(map[string]interface{})

	userID, _ := from["id"].(float64)
	chatID, _ := chat["id"].(float64)

	log.Printf("Received message: %s from user %d", text, int64(userID))

	// Обработка команд бота
	switch text {
	case "/start":
		role := "guest"
		if b.getUserRole != nil {
			role = b.getUserRole(int64(userID))
		}
		b.sendMainMenu(int64(chatID), role)
	case "/help":
		b.sendHelpMessage(int64(chatID))
	case "/app":
		b.sendAppLink(int64(chatID))
	case "/info":
		b.sendTeacherInfo(int64(chatID))
	default:
		if strings.HasPrefix(text, "/add_student") {
			b.handleAddStudent(int64(userID), text)
			return
		}
		// Проверяем, есть ли медиафайлы в сообщении
		if b.hasMediaFiles(message) {
			b.handleMediaMessage(message)
		} else {
			b.SendMessage(int64(chatID), "Используйте команду /start для начала работы с ботом.")
		}
	}
}

func (b *Bot) handleAddStudent(teacherTelegramID int64, text string) {
	if b.assignStudent == nil {
		b.SendMessage(teacherTelegramID, "Функция назначения ученика недоступна")
		return
	}
	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.SendMessage(teacherTelegramID, "Формат: /add_student @username|telegram_id [класс] [предметы]")
		return
	}
	var tgID *int64
	uname := ""
	// попытка распознать ID
	if id, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
		tgID = &id
	} else {
		uname = strings.TrimPrefix(parts[1], "@")
	}
	var grade *int
	if len(parts) >= 3 {
		if g, err := strconv.Atoi(parts[2]); err == nil {
			grade = &g
		}
	}
	subjects := ""
	if len(parts) >= 4 {
		subjects = strings.Join(parts[3:], " ")
	}
	if err := b.assignStudent(teacherTelegramID, tgID, uname, grade, subjects); err != nil {
		b.SendMessage(teacherTelegramID, fmt.Sprintf("Не удалось назначить ученика: %v", err))
		return
	}
	b.SendMessage(teacherTelegramID, "✅ Ученик назначен")
}

// sendMainMenu показывает главное меню по роли
func (b *Bot) sendMainMenu(chatID int64, role string) {
	var rows [][]tgbotapi.InlineKeyboardButton
	if role == "teacher" {
		rows = [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonURL("🔔 Уведомления", "https://edubot-0g05.onrender.com/teacher-dashboard")},
			{tgbotapi.NewInlineKeyboardButtonURL("👥 Ученики", "https://edubot-0g05.onrender.com/teacher-students")},
			{tgbotapi.NewInlineKeyboardButtonURL("👨‍👩‍👧 Группы (в приложении)", "https://edubot-0g05.onrender.com/teacher-groups")},
			{tgbotapi.NewInlineKeyboardButtonData("📋 Группы (в боте)", "show_groups")},
			{tgbotapi.NewInlineKeyboardButtonURL("📝 Задать ДЗ", "https://edubot-0g05.onrender.com/teacher-assignments")},
			{tgbotapi.NewInlineKeyboardButtonURL("✅ Проверка ДЗ", "https://edubot-0g05.onrender.com/teacher-submissions")},
			{tgbotapi.NewInlineKeyboardButtonURL("📚 Материалы", "https://edubot-0g05.onrender.com/teacher-content")},
		}
	} else if role == "student" {
		rows = [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonURL("📋 Мои задания", "https://edubot-0g05.onrender.com/student-dashboard")},
			{tgbotapi.NewInlineKeyboardButtonURL("📤 Сдать ДЗ", "https://edubot-0g05.onrender.com/student-dashboard")},
			{tgbotapi.NewInlineKeyboardButtonURL("❓ Помощь", "https://edubot-0g05.onrender.com")},
		}
	} else {
		rows = [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonURL("🚀 Открыть приложение", "https://edubot-0g05.onrender.com")},
			{tgbotapi.NewInlineKeyboardButtonData("ℹ️ Помощь", "help")},
		}
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	msg.ReplyMarkup = kb
	_, _ = b.api.Send(msg)
}

// processCallbackQuery обрабатывает инлайн-кнопки
func (b *Bot) processCallbackQuery(cb map[string]interface{}) {
	data, _ := cb["data"].(string)
	from, _ := cb["from"].(map[string]interface{})
	userID, _ := from["id"].(float64)
	message, _ := cb["message"].(map[string]interface{})
	chat, _ := message["chat"].(map[string]interface{})
	chatID, _ := chat["id"].(float64)

	switch data {
	case "help":
		b.sendHelpMessage(int64(chatID))
	case "show_groups":
		b.renderGroupsList(int64(chatID), int64(userID))
	default:
		// no-op
	}
}

func (b *Bot) renderGroupsList(chatID, teacherTelegramID int64) {
	if b.listGroups == nil {
		b.SendMessage(chatID, "Список групп недоступен")
		return
	}
	groups, err := b.listGroups(teacherTelegramID)
	if err != nil || len(groups) == 0 {
		b.SendMessage(chatID, "Группы не найдены. Создайте их в приложении.")
		return
	}
	// Рисуем до 10 кнопок; для простоты без пагинации
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, g := range groups {
		if i >= 10 {
			break
		}
		// Кнопка открывает страницу группы в приложении
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonURL("👥 "+g.Name, "https://edubot-0g05.onrender.com/teacher-groups"),
		})
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "/start")})
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "Ваши группы:")
	msg.ReplyMarkup = kb
	_, _ = b.api.Send(msg)
}

// hasMediaFiles проверяет, содержит ли сообщение медиафайлы
func (b *Bot) hasMediaFiles(message map[string]interface{}) bool {
	// Проверяем различные типы медиафайлов
	_, hasPhoto := message["photo"]
	_, hasVideo := message["video"]
	_, hasAudio := message["audio"]
	_, hasDocument := message["document"]
	_, hasVoice := message["voice"]

	return hasPhoto || hasVideo || hasAudio || hasDocument || hasVoice
}

// handleMediaMessage обрабатывает сообщения с медиафайлами
func (b *Bot) handleMediaMessage(message map[string]interface{}) {
	from, _ := message["from"].(map[string]interface{})
	chat, _ := message["chat"].(map[string]interface{})

	userID, _ := from["id"].(float64)
	chatID, _ := chat["id"].(float64)

	log.Printf("Received media message from user %d", int64(userID))

	// Здесь нужно будет интегрировать с MediaService
	// Пока просто отправляем подтверждение
	b.SendMessage(int64(chatID), "📎 Медиафайл получен! Спасибо за отправку.")
}

// duplicate callback handler removed (используется новая версия выше)

// sendWelcomeMessage отправляет приветственное сообщение с кнопкой
func (b *Bot) sendWelcomeMessage(chatID int64) error {
	text := `👋 Привет! Меня зовут Саша.

🎓 Я преподаватель физики и математики с 5-летним опытом подготовки к ЕГЭ.

📚 Чтобы познакомиться поближе, можешь перейти в приложение и узнать обо мне, моих методах обучения и записаться на пробное занятие.

🚀 Начнем путь к успешной сдаче ЕГЭ вместе!

💡 <b>Быстрая навигация:</b>
• Используй /start для возврата в главное меню
• Используй /help для получения помощи
• Нажми кнопку ниже для перехода в приложение`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Создаем клавиатуру с кнопкой "Открыть приложение"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🚀 Открыть приложение", "https://edubot-0g05.onrender.com"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Помощь", "help"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send welcome message: %w", err)
	}
	return nil
}

// sendHelpMessage отправляет сообщение с помощью
func (b *Bot) sendHelpMessage(chatID int64) error {
	text := `ℹ️ <b>Помощь по использованию EduBot</b>

🎯 <b>Основные функции:</b>
• 📝 Запись на пробные занятия
• 📚 Просмотр образовательных материалов
• 📋 Получение заданий и их выполнение
• 📊 Отслеживание прогресса обучения

🚀 <b>Доступные команды:</b>
• /start - Начать работу с ботом
• /help - Получить помощь
• /app - Открыть приложение
• /info - Информация о преподавателе

📱 <b>Как начать:</b>
1. Нажмите кнопку "Открыть приложение"
2. Заполните форму записи на пробное занятие
3. Дождитесь связи от преподавателя

❓ <b>Вопросы?</b>
Напишите преподавателю через приложение или используйте команду /start`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Создаем клавиатуру с кнопками
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🚀 Открыть приложение", "https://edubot-0g05.onrender.com"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨‍🏫 О преподавателе", "info"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главная", "start"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send help message: %w", err)
	}
	return nil
}

// sendAppLink отправляет ссылку на приложение
func (b *Bot) sendAppLink(chatID int64) error {
	text := `📱 <b>Открыть приложение EduBot</b>

🚀 Переходите в приложение для:
• Записи на пробное занятие
• Просмотра образовательных материалов
• Отслеживания прогресса обучения

Нажмите кнопку ниже для перехода в приложение!`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Создаем клавиатуру с кнопкой "Открыть приложение"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🚀 Открыть приложение", "https://edubot-0g05.onrender.com"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главная", "start"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send app link: %w", err)
	}
	return nil
}

// sendTeacherInfo отправляет информацию о преподавателе
func (b *Bot) sendTeacherInfo(chatID int64) error {
	text := `👨‍🏫 <b>Информация о преподавателе</b>

🎓 <b>Александр Пугачев</b>
• Преподаватель физики и математики
• 5 лет опыта подготовки к ЕГЭ
• Средний балл учеников: 85+

📚 <b>Специализация:</b>
• Физика (ЕГЭ)
• Профильная математика (ЕГЭ)
• Подготовка к олимпиадам

🏆 <b>Достижения:</b>
• Более 100 успешно подготовленных учеников
• Средний балл ЕГЭ: 85+
• Ученики поступают в ведущие вузы

💬 <b>Контакты:</b>
Telegram: @pugach3

🚀 Хотите начать обучение? Переходите в приложение!`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Создаем клавиатуру с кнопками
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🚀 Открыть приложение", "https://edubot-0g05.onrender.com"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главная", "start"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send teacher info: %w", err)
	}
	return nil
}

// (удалено дублирующееся определение SendAssignmentNotification)

// SendAssignmentCompletedNotification отправляет уведомление о выполненном задании
func (b *Bot) SendAssignmentCompletedNotification(chatID int64, title, subject, deadline string) error {
	text := fmt.Sprintf(`✅ <b>Задание выполнено!</b>

📋 <b>Название:</b> %s
📖 <b>Предмет:</b> %s
⏰ <b>Дедлайн был:</b> %s

👨‍🎓 Ученик выполнил задание. Проверьте работу в приложении!`, title, subject, deadline)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Создаем клавиатуру с кнопкой "Открыть приложение"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📝 Проверить задание", "https://edubot-0g05.onrender.com"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send assignment completed notification: %w", err)
	}
	return nil
}

// SendCommentNotification отправляет уведомление о новом комментарии
func (b *Bot) SendCommentNotification(chatID int64, content, title, subject string) error {
	text := fmt.Sprintf(`💬 <b>Новый комментарий к заданию!</b>

📋 <b>Задание:</b> %s
📖 <b>Предмет:</b> %s

💭 <b>Комментарий:</b>
%s

🚀 Переходите в приложение для просмотра!`, title, subject, content)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Создаем клавиатуру с кнопкой "Открыть приложение"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("💬 Посмотреть комментарий", "https://edubot-0g05.onrender.com"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send comment notification: %w", err)
	}
	return nil
}

// GetFilePath получает путь к файлу в Telegram
func (b *Bot) GetFilePath(fileID string) (string, error) {
	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.token, file.FilePath)
	return fileURL, nil
}

// SendMediaUploadInstructions отправляет инструкции по загрузке медиа
func (b *Bot) SendMediaUploadInstructions(chatID int64, mediaType string) error {
	var text string
	var keyboard tgbotapi.InlineKeyboardMarkup

	switch mediaType {
	case "welcome_video":
		text = `🎬 <b>Загрузка приветственного ролика</b>

Отправьте видео-файл, который будет отображаться на главной странице приложения.

<b>Требования:</b>
• Формат: MP4, MOV, AVI
• Размер: до 50 МБ
• Длительность: до 5 минут
• Качество: HD (720p) или выше

Просто отправьте видео-файл в этот чат!`

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("📱 Открыть приложение", "https://edubot-0g05.onrender.com"),
			),
		)

	case "homework":
		text = `📝 <b>Сдача домашнего задания</b>

Отправьте файлы с решением задания:
• Фото решений
• Видео с объяснением
• Документы (PDF, DOC)
• Аудио-комментарии

<b>Можно отправить несколько файлов!</b>`

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("📚 Мои задания", "https://edubot-0g05.onrender.com"),
			),
		)

	case "feedback":
		text = `🎯 <b>Запись фидбэка для ученика</b>

Запишите ваш отзыв о выполненном задании:
• Голосовое сообщение
• Видео с разбором ошибок
• Документ с комментариями
• Фото с пометками

Это поможет ученику лучше понять материал!`

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("👨‍🏫 Панель учителя", "https://edubot-0g05.onrender.com"),
			),
		)

	default:
		text = `📎 <b>Загрузка медиафайла</b>

Отправьте файл в этот чат для загрузки в приложение.

<b>Поддерживаемые форматы:</b>
• Видео: MP4, MOV, AVI
• Аудио: MP3, WAV, OGG
• Документы: PDF, DOC, DOCX
• Изображения: JPG, PNG, GIF`
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send media upload instructions: %w", err)
	}
	return nil
}

// SendMediaUploadSuccess отправляет уведомление об успешной загрузке
func (b *Bot) SendMediaUploadSuccess(chatID int64, mediaType, fileName string) error {
	var text string

	switch mediaType {
	case "welcome_video":
		text = fmt.Sprintf(`✅ <b>Приветственный ролик загружен!</b>

📹 <b>Файл:</b> %s

Ролик теперь отображается на главной странице приложения.`, fileName)

	case "homework":
		text = fmt.Sprintf(`✅ <b>Домашнее задание сдано!</b>

📎 <b>Файл:</b> %s

Ваше решение отправлено учителю на проверку.`, fileName)

	case "feedback":
		text = fmt.Sprintf(`✅ <b>Фидбэк записан!</b>

📎 <b>Файл:</b> %s

Ваш отзыв отправлен ученику.`, fileName)

	default:
		text = fmt.Sprintf(`✅ <b>Медиафайл загружен!</b>

📎 <b>Файл:</b> %s

Файл успешно добавлен в приложение.`, fileName)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send media upload success: %w", err)
	}
	return nil
}

// SendMediaUploadError отправляет уведомление об ошибке загрузки
func (b *Bot) SendMediaUploadError(chatID int64, errorMsg string) error {
	text := fmt.Sprintf(`❌ <b>Ошибка загрузки медиафайла</b>

%s

Попробуйте еще раз или обратитесь в поддержку.`, errorMsg)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send media upload error: %w", err)
	}
	return nil
}

// HandleMediaUpload обрабатывает загрузку медиафайлов
func (b *Bot) HandleMediaUpload(update tgbotapi.Update, mediaService interface{}) error {
	var fileName string
	var mediaType string

	// Определяем тип медиафайла и извлекаем информацию
	if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
		// Изображение
		photo := update.Message.Photo[len(update.Message.Photo)-1] // Берем самое большое
		fileName = fmt.Sprintf("image_%d.jpg", photo.FileSize)
		mediaType = "image"
	} else if update.Message.Video != nil {
		// Видео
		fileName = update.Message.Video.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("video_%d.mp4", update.Message.Video.FileSize)
		}
		mediaType = "video"
	} else if update.Message.Audio != nil {
		// Аудио
		fileName = update.Message.Audio.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("audio_%d.mp3", update.Message.Audio.FileSize)
		}
		mediaType = "audio"
	} else if update.Message.Document != nil {
		// Документ
		fileName = update.Message.Document.FileName
		mediaType = "document"
	} else if update.Message.Voice != nil {
		// Голосовое сообщение
		fileName = fmt.Sprintf("voice_%d.ogg", update.Message.Voice.FileSize)
		mediaType = "audio"
	} else {
		return fmt.Errorf("unsupported media type")
	}

	chatID := update.Message.Chat.ID

	// Здесь нужно будет интегрировать с MediaService
	// Пока просто отправляем подтверждение
	err := b.SendMediaUploadSuccess(chatID, mediaType, fileName)
	if err != nil {
		return fmt.Errorf("failed to send success notification: %w", err)
	}

	return nil
}

// HandleWelcomeVideoUpload обрабатывает загрузку приветственного ролика
func (b *Bot) HandleWelcomeVideoUpload(update tgbotapi.Update, mediaService interface{}) error {
	if update.Message.Video == nil {
		return fmt.Errorf("expected video file")
	}

	// Проверяем, что пользователь является учителем
	// Здесь нужно будет добавить проверку роли пользователя

	return b.HandleMediaUpload(update, mediaService)
}

// HandleHomeworkSubmission обрабатывает сдачу домашнего задания
func (b *Bot) HandleHomeworkSubmission(update tgbotapi.Update, mediaService interface{}) error {
	// Проверяем, что пользователь является учеником
	// Здесь нужно будет добавить проверку роли пользователя

	return b.HandleMediaUpload(update, mediaService)
}

// HandleTeacherFeedback обрабатывает запись фидбэка учителем
func (b *Bot) HandleTeacherFeedback(update tgbotapi.Update, mediaService interface{}) error {
	// Проверяем, что пользователь является учителем
	// Здесь нужно будет добавить проверку роли пользователя

	return b.HandleMediaUpload(update, mediaService)
}

// SendFeedbackNotification отправляет уведомление о фидбэке от учителя
func (b *Bot) SendFeedbackNotification(userTelegramID int64, assignmentTitle, subject, grade, comments string) {
	gradeText := grade
	if grade == "needs_revision" {
		gradeText = "на доработку"
	}

	message := fmt.Sprintf(
		"📝 *Ваше задание проверено!*\n\n"+
			"📚 Задание: %s\n"+
			"📖 Предмет: %s\n"+
			"⭐ Оценка: %s\n\n",
		assignmentTitle, subject, gradeText,
	)

	if comments != "" {
		message += fmt.Sprintf("💬 Комментарий учителя:\n%s\n\n", comments)
	}

	if grade == "needs_revision" {
		message += "🔄 Пожалуйста, доработайте задание и отправьте заново."
	} else {
		message += "✅ Отличная работа! Продолжайте в том же духе."
	}

	b.SendMessage(userTelegramID, message)
	log.Printf("Feedback notification sent to user %d for assignment %s", userTelegramID, assignmentTitle)
}
