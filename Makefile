# Makefile для EduBot

# Переменные
APP_NAME=edubot
BINARY_NAME=edubot
BUILD_DIR=build
DOCKER_IMAGE=edubot:latest

# Цвета для вывода
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build run clean test deps install docker-build docker-run dev setup

# Помощь
help:
	@echo "$(BLUE)EduBot - Образовательная платформа$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@echo "  $(GREEN)build$(NC)        - Собрать приложение"
	@echo "  $(GREEN)run$(NC)          - Запустить приложение"
	@echo "  $(GREEN)dev$(NC)          - Запустить в режиме разработки"
	@echo "  $(GREEN)test$(NC)         - Запустить тесты"
	@echo "  $(GREEN)clean$(NC)        - Очистить собранные файлы"
	@echo "  $(GREEN)deps$(NC)         - Установить зависимости"
	@echo "  $(GREEN)install$(NC)      - Установить приложение"
	@echo "  $(GREEN)setup$(NC)        - Первоначальная настройка"
	@echo "  $(GREEN)docker-build$(NC) - Собрать Docker образ"
	@echo "  $(GREEN)docker-run$(NC)   - Запустить в Docker"
	@echo "  $(GREEN)lint$(NC)         - Проверить код линтером"
	@echo "  $(GREEN)format$(NC)       - Форматировать код"

# Установка зависимостей
deps:
	@echo "$(BLUE)Установка зависимостей...$(NC)"
	go mod download
	go mod tidy

# Первоначальная настройка
setup: deps
	@echo "$(BLUE)Первоначальная настройка проекта...$(NC)"
	@mkdir -p data uploads/web/static uploads/users
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Создание .env файла из примера...$(NC)"; \
		cp env.example .env; \
		echo "$(GREEN)Отредактируйте .env файл с вашими настройками$(NC)"; \
	fi
	@echo "$(GREEN)Настройка завершена!$(NC)"

# Сборка приложения
build: deps
	@echo "$(BLUE)Сборка приложения...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "$(GREEN)Приложение собрано: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Запуск приложения
run: build
	@echo "$(BLUE)Запуск приложения...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# Запуск в режиме разработки
dev:
	@echo "$(BLUE)Запуск в режиме разработки...$(NC)"
	@echo "$(YELLOW)Используйте Ctrl+C для остановки$(NC)"
	go run cmd/main.go

# Запуск тестов
test:
	@echo "$(BLUE)Запуск тестов...$(NC)"
	go test -v ./...

# Проверка кода линтером
lint:
	@echo "$(BLUE)Проверка кода линтером...$(NC)"
	golangci-lint run

# Форматирование кода
format:
	@echo "$(BLUE)Форматирование кода...$(NC)"
	go fmt ./...
	goimports -w .

# Установка приложения
install: build
	@echo "$(BLUE)Установка приложения...$(NC)"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)Приложение установлено в /usr/local/bin/$(NC)"

# Очистка
clean:
	@echo "$(BLUE)Очистка...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean
	@echo "$(GREEN)Очистка завершена$(NC)"

# Docker сборка
docker-build:
	@echo "$(BLUE)Сборка Docker образа...$(NC)"
	docker build -t $(DOCKER_IMAGE) .
	@echo "$(GREEN)Docker образ собран: $(DOCKER_IMAGE)$(NC)"

# Docker запуск
docker-run: docker-build
	@echo "$(BLUE)Запуск в Docker...$(NC)"
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

# Создание миграций базы данных
migrate:
	@echo "$(BLUE)Применение миграций...$(NC)"
	go run cmd/main.go migrate

# Создание резервной копии базы данных
backup:
	@echo "$(BLUE)Создание резервной копии...$(NC)"
	@mkdir -p backups
	cp data/edubot.db backups/edubot_$(shell date +%Y%m%d_%H%M%S).db
	@echo "$(GREEN)Резервная копия создана$(NC)"

# Восстановление из резервной копии
restore:
	@echo "$(BLUE)Восстановление из резервной копии...$(NC)"
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "$(RED)Укажите файл резервной копии: make restore BACKUP_FILE=backups/edubot_YYYYMMDD_HHMMSS.db$(NC)"; \
		exit 1; \
	fi
	cp $(BACKUP_FILE) data/edubot.db
	@echo "$(GREEN)База данных восстановлена$(NC)"

# Генерация документации
docs:
	@echo "$(BLUE)Генерация документации...$(NC)"
	godoc -http=:6060
	@echo "$(GREEN)Документация доступна по адресу: http://localhost:6060$(NC)"

# Проверка безопасности
security:
	@echo "$(BLUE)Проверка безопасности...$(NC)"
	gosec ./...

# Профилирование
profile:
	@echo "$(BLUE)Запуск с профилированием...$(NC)"
	go run cmd/main.go -cpuprofile=cpu.prof -memprofile=mem.prof

# Анализ покрытия тестами
coverage:
	@echo "$(BLUE)Анализ покрытия тестами...$(NC)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии сохранен в coverage.html$(NC)"

# Проверка зависимостей
check-deps:
	@echo "$(BLUE)Проверка зависимостей...$(NC)"
	go mod verify
	go list -m all

# Обновление зависимостей
update-deps:
	@echo "$(BLUE)Обновление зависимостей...$(NC)"
	go get -u ./...
	go mod tidy

# Создание релиза
release:
	@echo "$(BLUE)Создание релиза...$(NC)"
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Укажите версию: make release VERSION=v1.0.0$(NC)"; \
		exit 1; \
	fi
	@mkdir -p releases
	GOOS=linux GOARCH=amd64 go build -o releases/$(BINARY_NAME)-linux-amd64-$(VERSION) cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o releases/$(BINARY_NAME)-windows-amd64-$(VERSION).exe cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o releases/$(BINARY_NAME)-darwin-amd64-$(VERSION) cmd/main.go
	@echo "$(GREEN)Релизы созданы в папке releases/$(NC)"

# Статистика проекта
stats:
	@echo "$(BLUE)Статистика проекта:$(NC)"
	@echo "Строк кода Go: $(shell find . -name '*.go' | xargs wc -l | tail -1)"
	@echo "Размер проекта: $(shell du -sh . | cut -f1)"
	@echo "Количество файлов: $(shell find . -type f | wc -l)"

# Информация о проекте
info:
	@echo "$(BLUE)Информация о проекте:$(NC)"
	@echo "Название: $(APP_NAME)"
	@echo "Версия Go: $(shell go version)"
	@echo "Архитектура: $(shell uname -m)"
	@echo "ОС: $(shell uname -s)"
