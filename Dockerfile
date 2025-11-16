# Мультистейдж сборка для минимизации размера образа
FROM golang:1.25-alpine AS builder

# Установка необходимых зависимостей для сборки
RUN apk add --no-cache git make

# Рабочая директория
WORKDIR /app

# Копирование go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование всего исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/api ./cmd/api/main.go

# Финальный образ с минимальным размером
FROM alpine:latest

# Установка CA сертификатов для HTTPS и timezone data
RUN apk --no-cache add ca-certificates tzdata

# Создание непривилегированного пользователя
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Копирование бинаря из builder stage
COPY --from=builder /app/bin/api /app/api

# Копирование TLS сертификатов (если они есть)
COPY --from=builder /app/certs /app/certs

# Создание директории для логов
RUN mkdir -p /app/logs && chown -R appuser:appuser /app

# Переключение на непривилегированного пользователя
USER appuser

# Открытие порта
EXPOSE 8443

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --no-check-certificate https://localhost:8443/health || exit 1

# Запуск приложения
CMD ["/app/api"]
