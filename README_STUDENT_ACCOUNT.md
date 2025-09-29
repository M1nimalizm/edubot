# Student Personal Account (ЛК ученика) - Implementation Guide

## 🎯 Overview
Реализован полнофункциональный личный кабинет ученика с интеграцией в Telegram Mini App, включающий:
- Управление заданиями и их выполнение
- Чат с преподавателем
- Отслеживание прогресса
- Систему уведомлений

## 🏗️ Architecture

### Models
- **AssignmentTarget**: Индивидуальные задания для студентов
- **Submission**: Отправки решений студентами
- **Feedback**: Обратная связь от преподавателей
- **ChatThread/Message**: Система чатов
- **Notification**: Уведомления

### Services
- **AssignmentService**: Управление заданиями и их распределением
- **SubmissionService**: Обработка отправок и черновиков
- **GradingService**: Система оценивания и фидбэка
- **ChatService**: Чат между студентами и преподавателями
- **NotificationService**: Уведомления и напоминания

### API Endpoints

#### Student API (`/api/student/`)
- `GET /assignments` - Список заданий студента
- `GET /assignments/:id` - Детали задания
- `POST /assignments/:id/submit` - Отправка решения
- `POST /assignments/:id/draft` - Сохранение черновика
- `GET /assignments/:id/draft` - Получение черновика
- `GET /progress` - Статистика прогресса
- `GET /notifications` - Уведомления
- `POST /notifications/:id/read` - Отметить как прочитанное

#### Chat API (`/api/chat/`)
- `GET /threads` - Список чатов
- `GET /threads/:id` - Детали чата
- `POST /threads/student-teacher` - Создать чат с преподавателем
- `POST /threads/group` - Создать групповой чат
- `GET /threads/:id/messages` - Сообщения чата
- `POST /threads/:id/messages` - Отправить сообщение
- `PUT /messages/:id` - Редактировать сообщение
- `DELETE /messages/:id` - Удалить сообщение
- `POST /threads/:id/read` - Отметить как прочитанное

#### Teacher Inbox API (`/api/teacher/inbox/`)
- `GET /inbox` - Входящие для проверки
- `GET /inbox/:id` - Детали задания для оценки
- `POST /inbox/:id/grade` - Оценить задание
- `GET /assignments` - Список заданий преподавателя
- `POST /assignments` - Создать задание
- `GET /statistics` - Статистика
- `GET /notifications` - Уведомления преподавателя

## 🤖 Telegram Bot Integration

### Bot Menus
- **Student Menu**: Мои задания, Сдать ДЗ, Чат с учителем, Мой прогресс
- **Teacher Menu**: Уведомления, Ученики, Группы, Задать ДЗ, Проверка ДЗ, Записать фидбэк

### Bot Modes
- **Student Submit Mode**: Режим отправки файлов с решениями
- **Teacher Feedback Mode**: Режим записи фидбэка для студентов

### Deep Links
- `https://edubot-0g05.onrender.com/app/student-dashboard?assignment=ID` - Открыть конкретное задание
- `https://edubot-0g05.onrender.com/app/student-progress?subject=physics` - Прогресс по предмету
- `https://edubot-0g05.onrender.com/app/student-chat?teacher=name` - Чат с преподавателем

## 📱 Mini App Features

### Navigation
- Сохранение состояния приложения
- Кнопка "Назад" с историей навигации
- Deep linking для прямого перехода к разделам

### Student Dashboard
- Список заданий с фильтрацией по статусу
- Статистика выполнения
- Быстрый доступ к чату и прогрессу

### Assignment Detail
- Полная информация о задании
- Отправка решений с медиафайлами
- Просмотр фидбэка преподавателя

### Chat Interface
- Личные чаты с преподавателями
- Групповые чаты
- Отправка медиафайлов
- Индикаторы непрочитанных сообщений

### Progress Tracking
- Общая статистика по всем предметам
- Детальная статистика по предметам
- Графики выполнения заданий

## 🔐 Security & RBAC

### Authentication
- JWT токены в заголовках и cookies
- Telegram WebApp авторизация
- Валидация токенов на каждом запросе

### Authorization
- **Student Routes**: Только для студентов
- **Teacher Routes**: Только для преподавателей
- **Shared Routes**: Для студентов и преподавателей
- **Public Routes**: Доступны всем

### Middleware
- `AuthMiddleware`: Проверка авторизации
- `StudentOnlyMiddleware`: Только студенты
- `TeacherOnlyMiddleware`: Только преподаватели
- `RequireRoles`: Гибкая проверка ролей

## 🚀 Deployment

### Environment Variables
```bash
PORT=8080
BASE_URL=https://edubot-0g05.onrender.com
JWT_SECRET=your-secret-key
DATABASE_URL=postgresql://...
TELEGRAM_BOT_TOKEN=your-bot-token
TELEGRAM_WEBHOOK_URL=https://edubot-0g05.onrender.com/webhook
DISABLE_SITE=true
UPLOAD_PATH=./uploads
TEACHER_TELEGRAM_IDS=123456789,987654321
```

### Database
- PostgreSQL для продакшена
- SQLite для разработки
- Автоматические миграции через GORM

## 📊 Monitoring & Logs
- Логирование всех операций
- Мониторинг ошибок
- Статистика использования API

## 🔧 Development

### Building
```bash
go build -o edubot ./cmd
```

### Running
```bash
./edubot
```

### Testing
- Все API endpoints протестированы
- Интеграция с Telegram Bot работает
- Mini App функциональность проверена
