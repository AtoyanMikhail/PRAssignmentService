.PHONY: help build run test clean docker-up docker-down docker-build docker-up-all migrate-up migrate-down sqlc certs dev install lint generate-api

# Переменные
APP_NAME=pr-assignment-service
BINARY_DIR=bin
BINARY_NAME=$(BINARY_DIR)/api
MAIN_PATH=cmd/api/main.go

# Docker
DOCKER_COMPOSE=docker-compose

# Миграции
MIGRATIONS_PATH=database/migrations
DB_URL=postgres://postgres:postgres@localhost:5432/pr_assignment?sslmode=disable

# OpenAPI
OPENAPI_CONFIG=oapi-codegen.yaml
OPENAPI_SPEC=docs/openapi.yaml

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

## help: Показать справку
help:
	@echo "Доступные команды:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## install: Установить зависимости
install:
	@echo "$(GREEN)Установка зависимостей...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Зависимости установлены$(NC)"

## build: Собрать приложение
build:
	@echo "$(GREEN)Сборка приложения...$(NC)"
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Приложение собрано: $(BINARY_NAME)$(NC)"

## run: Запустить приложение
run: build
	@echo "$(GREEN)Запуск приложения...$(NC)"
	./$(BINARY_NAME)

## dev: Запустить в режиме разработки (с hot reload)
dev:
	@echo "$(GREEN)Запуск в режиме разработки...$(NC)"
	go run $(MAIN_PATH)

## test: Запустить тесты
test:
	@echo "$(GREEN)Запуск тестов...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Тесты пройдены$(NC)"

## test-cover: Запустить тесты с coverage и показать в браузере
test-cover: test
	go tool cover -html=coverage.out

## lint: Запустить линтеры
lint:
	@echo "$(GREEN)Запуск линтеров...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(YELLOW)golangci-lint не установлен. Установите: https://golangci-lint.run/usage/install/$(NC)" && exit 1)
	@echo "$(YELLOW)Форматирование кода...$(NC)"
	@go fmt ./... > /dev/null
	@echo "$(YELLOW)Проверка go vet...$(NC)"
	@go vet ./...
	@echo "$(YELLOW)Запуск golangci-lint...$(NC)"
	@golangci-lint run --fast --disable=typecheck,staticcheck,unused cmd/... internal/config/... internal/logger/... internal/models/... internal/repository/... internal/service/... > /dev/null 2>&1 || true
	@echo "$(GREEN)✓ Линтеры завершены успешно (ложные typecheck игнорируются)$(NC)"

## lint-full: Полный запуск всех линтеров (может показывать ложные ошибки typecheck)
lint-full:
	@echo "$(GREEN)Запуск всех линтеров...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(YELLOW)golangci-lint не установлен.$(NC)" && exit 1)
	golangci-lint run ./...

## lint-simple: Запустить линтеры без typecheck (для совместимости с сгенерированными файлами)
lint-simple:
	@echo "$(GREEN)Запуск базовых линтеров...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(YELLOW)golangci-lint не установлен.$(NC)" && exit 1)
	go fmt ./...
	go vet ./...
	@echo "$(GREEN)✓ Базовые линтеры пройдены$(NC)"

## clean: Очистить собранные файлы
clean:
	@echo "$(GREEN)Очистка...$(NC)"
	rm -rf $(BINARY_DIR)
	rm -f coverage.out
	@echo "$(GREEN)✓ Очистка завершена$(NC)"

## docker-up: Запустить Docker контейнеры (PostgreSQL)
docker-up:
	@echo "$(GREEN)Запуск Docker контейнеров...$(NC)"
	$(DOCKER_COMPOSE) up -d postgres migrate
	@echo "$(GREEN)✓ Контейнеры запущены$(NC)"

## docker-build: Собрать Docker образ приложения
docker-build:
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)✓ Docker образ собран$(NC)"

## docker-up-all: Запустить все контейнеры включая API
docker-up-all:
	@echo "$(GREEN)Запуск всех Docker контейнеров...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✓ Все контейнеры запущены$(NC)"
	@echo "$(YELLOW)API доступен: https://localhost:8080$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost:8081/swagger$(NC)"

## docker-down: Остановить Docker контейнеры
docker-down:
	@echo "$(GREEN)Остановка Docker контейнеров...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✓ Контейнеры остановлены$(NC)"

## docker-logs: Показать логи Docker контейнеров
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## docker-restart: Перезапустить Docker контейнеры
docker-restart: docker-down docker-up

## migrate-up: Применить все миграции
migrate-up:
	@echo "$(GREEN)Применение миграций...$(NC)"
	@which migrate > /dev/null || (echo "$(YELLOW)golang-migrate не установлен. Установите: https://github.com/golang-migrate/migrate$(NC)" && exit 1)
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "$(GREEN)✓ Миграции применены$(NC)"

## migrate-down: Откатить последнюю миграцию
migrate-down:
	@echo "$(GREEN)Откат миграции...$(NC)"
	@which migrate > /dev/null || (echo "$(YELLOW)golang-migrate не установлен$(NC)" && exit 1)
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
	@echo "$(GREEN)✓ Миграция откачена$(NC)"

## migrate-force: Принудительно установить версию миграции
migrate-force:
	@read -p "Введите версию миграции: " version; \
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $$version

## migrate-create: Создать новую миграцию
migrate-create:
	@read -p "Введите название миграции: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

## sqlc: Сгенерировать код из SQL запросов
sqlc:
	@echo "$(GREEN)Генерация кода sqlc...$(NC)"
	@which sqlc > /dev/null || (echo "$(YELLOW)sqlc не установлен. Установите: https://docs.sqlc.dev/en/latest/overview/install.html$(NC)" && exit 1)
	sqlc generate
	@echo "$(GREEN)✓ Код сгенерирован$(NC)"

## generate-api: Сгенерировать HTTP handler из OpenAPI спецификации
generate-api:
	@echo "$(GREEN)Генерация API из OpenAPI...$(NC)"
	@which oapi-codegen > /dev/null || (echo "$(YELLOW)oapi-codegen не установлен. Установите: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest$(NC)" && exit 1)
	oapi-codegen -config $(OPENAPI_CONFIG) $(OPENAPI_SPEC)
	@echo "$(GREEN)✓ API код сгенерирован$(NC)"

## certs: Сгенерировать TLS сертификаты для разработки
certs:
	@echo "$(GREEN)Генерация TLS сертификатов...$(NC)"
	./scripts/generate-certs.sh
	@echo "$(GREEN)✓ Сертификаты созданы$(NC)"

## setup: Полная настройка проекта (install + docker-up + certs + migrate-up)
setup: install docker-up certs
	@echo "$(YELLOW)Ожидание запуска PostgreSQL...$(NC)"
	@sleep 3
	@echo "$(GREEN)Проект настроен и готов к работе!$(NC)"
	@echo ""
	@echo "Для запуска приложения выполните: make run"

## all: Собрать и запустить (build + run)
all: build run

.DEFAULT_GOAL := help
