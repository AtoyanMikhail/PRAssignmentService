package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/api"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/config"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/logger"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/repository"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Инициализация логгера
	logger.Init(os.Stdout)
	log := logger.Get()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Info("Configuration loaded successfully")

	// Создание контекста с отменой для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Подключение к базе данных
	log.Info("Connecting to database...")
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Unable to parse database config: %v", err)
	}

	// Настройка пула соединений из конфига
	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.Database.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Info("Successfully connected to database")

	// Инициализация Store с репозиториями
	store := repository.NewStore(pool)
	log.Info("Repository layer initialized")

	// Инициализация сервисного слоя
	services := service.NewServices(store)
	log.Info("Service layer initialized")

	// Инициализация HTTP handler
	handler := api.NewHandler(services)

	// Настройка Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Регистрация middleware
	router.Use(api.CORSMiddleware())
	router.Use(api.LoggingMiddleware(log))

	// Регистрация маршрутов из OpenAPI
	api.RegisterHandlers(router, handler)

	log.Info("HTTP server initialized")

	// Запуск HTTPS server в отдельной горутине
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	go func() {
		log.Infof("Starting HTTPS server on %s:%s...", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}()

	log.Info("Application started. Press Ctrl+C to exit...")

	// Ожидание сигнала завершения
	<-ctx.Done()

	// Восстановление дефолтного поведения сигналов для возможности принудительного завершения
	stop()
	log.Info("Shutting down gracefully...")

	// Создание контекста с таймаутом для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Остановка HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	// Ожидание завершения всех операций или таймаута
	<-shutdownCtx.Done()

	// Флаш буферов логгера перед выходом
	_ = log.Sync()
	log.Info("Application stopped")
}
