# Multi-stage build для EduBot

# Этап сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata

# Создаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o edubot cmd/main.go

# Этап продакшена
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя для безопасности
RUN adduser -D -s /bin/sh appuser

# Создаем необходимые директории
RUN mkdir -p /app/data /app/uploads/web/static /app/uploads/users && \
    chown -R appuser:appuser /app

# Переключаемся на пользователя
USER appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранное приложение
COPY --from=builder /app/edubot .

# Копируем веб-файлы
COPY --from=builder /app/web ./web

# Открываем порт
EXPOSE 8080

# Переменные окружения по умолчанию
ENV PORT=8080
ENV HOST=0.0.0.0
ENV DB_PATH=/app/data/edubot.db
ENV UPLOAD_PATH=/app/uploads

# Команда запуска
CMD ["./edubot"]
