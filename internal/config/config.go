package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config содержит всю конфигурацию приложения
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Logger   LoggerConfig
}

// DatabaseConfig содержит настройки базы данных
type DatabaseConfig struct {
	Host              string
	Port              string
	User              string
	Password          string
	DBName            string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// ServerConfig содержит настройки HTTP сервера
type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	TLSEnabled      bool
	TLSCertFile     string
	TLSKeyFile      string
}

// LoggerConfig содержит настройки логгера
type LoggerConfig struct {
	Level  string
	Format string // json или console
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:              getEnv("DB_HOST", "localhost"),
			Port:              getEnv("DB_PORT", "5432"),
			User:              getEnv("DB_USER", "postgres"),
			Password:          getEnv("DB_PASSWORD", "postgres"),
			DBName:            getEnv("DB_NAME", "pr_assignment"),
			SSLMode:           getEnv("DB_SSL_MODE", "disable"),
			MaxConns:          getEnvAsInt32("DB_MAX_CONNS", 25),
			MinConns:          getEnvAsInt32("DB_MIN_CONNS", 5),
			MaxConnLifetime:   getEnvAsDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:   getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getEnvAsDuration("DB_HEALTH_CHECK_PERIOD", time.Minute),
		},
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8443"),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
			TLSEnabled:      getEnvAsBool("SERVER_TLS_ENABLED", true),
			TLSCertFile:     getEnv("SERVER_TLS_CERT_FILE", "certs/server.crt"),
			TLSKeyFile:      getEnv("SERVER_TLS_KEY_FILE", "certs/server.key"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "console"),
		},
	}

	return cfg, nil
}

// DSN возвращает строку подключения к базе данных
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Address возвращает адрес сервера в формате host:port
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt32 возвращает значение переменной окружения как int32
func getEnvAsInt32(key string, defaultValue int32) int32 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int32(value)
}

// getEnvAsDuration возвращает значение переменной окружения как time.Duration
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsBool возвращает значение переменной окружения как bool
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
