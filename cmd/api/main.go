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
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(os.Stdout, cfg.Logger.Level, cfg.Logger.Format)
	log := logger.Get()

	log.Info("Configuration loaded successfully")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("Connecting to database...")
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Unable to parse database config: %v", err)
	}

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

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Info("Successfully connected to database")

	store := repository.NewStore(pool)
	log.Info("Repository layer initialized")

	services := service.NewServices(store)
	log.Info("Service layer initialized")

	handler := api.NewHandler(services)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(api.CORSMiddleware())
	router.Use(api.LoggingMiddleware(log))

	api.RegisterHandlers(router, handler)

	log.Info("HTTP server initialized")

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

	<-ctx.Done()

	stop()
	log.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	<-shutdownCtx.Done()

	_ = log.Sync()
	log.Info("Application stopped")
}
