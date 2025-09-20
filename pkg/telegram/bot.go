package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет Telegram бота
type Bot struct {
	api     *tgbotapi.BotAPI
	token   string
	webhook string
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
		// Проверяем, новый ли это пользователь
		firstName, _ := from["first_name"].(string)
		if firstName == "" {
			firstName = "друг"
		}
		
		// Отправляем персонализированное приветствие
		b.SendWelcomeToNewUser(int64(chatID), firstName)
	case "/help":
		b.sendHelpMessage(int64(chatID))
	case "/app":
		b.sendAppLink(int64(chatID))
	case "/info":
		b.sendTeacherInfo(int64(chatID))
	default:
		b.SendMessage(int64(chatID), "Используйте команду /start для начала работы с ботом.")
	}
}

// processCallbackQuery обрабатывает нажатия на inline-кнопки
func (b *Bot) processCallbackQuery(callbackQuery map[string]interface{}) {
	data, _ := callbackQuery["data"].(string)
	from, _ := callbackQuery["from"].(map[string]interface{})
	message, _ := callbackQuery["message"].(map[string]interface{})
	chat, _ := message["chat"].(map[string]interface{})

	userID, _ := from["id"].(float64)
	chatID, _ := chat["id"].(float64)
	callbackID, _ := callbackQuery["id"].(string)

	log.Printf("Received callback: %s from user %d", data, int64(userID))

	// Отвечаем на callback query
	callback := tgbotapi.NewCallback(callbackID, "")
	b.api.Request(callback)

	// Обрабатываем данные кнопки
	switch data {
	case "help":
		b.sendHelpMessage(int64(chatID))
	case "start":
		// Получаем имя пользователя для персонализации
		firstName, _ := from["first_name"].(string)
		if firstName == "" {
			firstName = "друг"
		}
		b.SendWelcomeToNewUser(int64(chatID), firstName)
	case "info":
		b.sendTeacherInfo(int64(chatID))
	default:
		b.SendMessage(int64(chatID), "Используйте команду /start для начала работы с ботом.")
	}
}

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
