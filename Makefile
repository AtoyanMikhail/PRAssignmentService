.PHONY: help build run test clean docker-up docker-down docker-build docker-up-all migrate-up migrate-down sqlc certs dev install lint generate-api

APP_NAME=pr-assignment-service
BINARY_DIR=bin
BINARY_NAME=$(BINARY_DIR)/api
MAIN_PATH=cmd/api/main.go

DOCKER_COMPOSE=docker-compose

MIGRATIONS_PATH=database/migrations
DB_URL=postgres://postgres:postgres@localhost:5432/pr_assignment?sslmode=disable

OPENAPI_CONFIG=oapi-codegen.yaml
OPENAPI_SPEC=docs/openapi.yaml

GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # 

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

## dev: Запустить в режиме разработки
dev:
	@echo "$(GREEN)Запуск в режиме разработки...$(NC)"
	go run $(MAIN_PATH)

## test: Запустить unit тесты
test:
	@echo "$(GREEN)Запуск unit тестов...$(NC)"
	go test -v -race -coverprofile=coverage.out ./internal/... ./cmd/...
	@echo "$(GREEN)✓ Unit тесты пройдены$(NC)"

## test-e2e: Запустить E2E тесты (требует запущенного сервиса)
test-e2e:
	@echo "$(GREEN)Запуск E2E тестов...$(NC)"
	@echo "$(YELLOW)Убедитесь, что сервис запущен (make docker-up-all)$(NC)"
	go test -v ./tests/e2e_test.go
	@echo "$(GREEN)✓ E2E тесты пройдены$(NC)"

## test-load: Запустить нагрузочные тесты (требует запущенного сервиса)
test-load:
	@echo "$(GREEN)Запуск нагрузочных тестов...$(NC)"
	@echo "$(YELLOW)Убедитесь, что сервис запущен (make docker-up-all)$(NC)"
	go test -v ./tests/load_test.go
	@echo "$(GREEN)✓ Нагрузочные тесты пройдены$(NC)"

## test-all: Запустить все тесты (unit + e2e + load)
test-all: test test-e2e test-load
	@echo "$(GREEN)✓ Все тесты пройдены$(NC)"

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
	@echo "$(GREEN)✓ Линтеры завершены успешно$(NC)"

## clean: Очистить собранные файлы
clean:
	@echo "$(GREEN)Очистка...$(NC)"
	rm -rf $(BINARY_DIR)
	rm -f coverage.out
	@echo "$(GREEN)✓ Очистка завершена$(NC)"

## docker-up-all: Запустить все контейнеры включая API
docker-up-all:
	@echo "$(GREEN)Запуск всех Docker контейнеров...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✓ Все контейнеры запущены$(NC)"
	@echo "$(YELLOW)API доступен: https://localhost:8080$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost:8081/swagger$(NC)"

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
