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
	text := fmt.Sprintf(`
🎓 <b>Новая заявка на пробное занятие!</b>

👤 <b>Имя:</b> %s
📚 <b>Класс:</b> %d
📖 <b>Предмет:</b> %s
⭐ <b>Уровень:</b> %d/5
📱 <b>Телефон:</b> %s

💬 <b>Комментарий:</b>
%s

🕐 <b>Время подачи:</b> %s
	`,
		requestData["name"],
		requestData["grade"],
		requestData["subject"],
		requestData["level"],
		requestData["phone"],
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
		b.SendMessage(int64(chatID), "Добро пожаловать в EduBot! Переходите в приложение для продолжения.")
	case "/help":
		b.SendMessage(int64(chatID), "Это бот для образовательной платформы EduBot. Используйте приложение для полного функционала.")
	default:
		b.SendMessage(int64(chatID), "Используйте приложение EduBot для взаимодействия с платформой.")
	}
}
