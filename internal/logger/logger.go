package logger

import (
	"io"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *Logger
	once     sync.Once
)

// Logger представляет синглтон логгер приложения
type Logger struct {
	zap *zap.Logger
}

// Init инициализирует глобальный логгер
func Init(writer io.Writer, level string, format string) {
	once.Do(func() {
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// Парсим уровень логирования
		var zapLevel zapcore.Level
		switch level {
		case "debug":
			zapLevel = zapcore.DebugLevel
		case "info":
			zapLevel = zapcore.InfoLevel
		case "warn", "warning":
			zapLevel = zapcore.WarnLevel
		case "error":
			zapLevel = zapcore.ErrorLevel
		case "fatal":
			zapLevel = zapcore.FatalLevel
		default:
			zapLevel = zapcore.InfoLevel // по умолчанию info
		}

		// Выбираем энкодер в зависимости от формата
		var encoder zapcore.Encoder
		switch format {
		case "json":
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		case "console":
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		default:
			encoder = zapcore.NewConsoleEncoder(encoderConfig) // по умолчанию console
		}

		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(writer),
			zapLevel,
		)

		instance = &Logger{
			zap: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
		}
	})
}

// Get возвращает экземпляр глобального логгера
func Get() *Logger {
	if instance == nil {
		panic("logger not initialized, call Init() first")
	}
	return instance
}

// Info логирует информационное сообщение
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Infof логирует форматированное информационное сообщение
func (l *Logger) Infof(format string, args ...interface{}) {
	l.zap.Sugar().Infof(format, args...)
}

// Warn логирует предупреждение
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Warnf логирует форматированное предупреждение
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zap.Sugar().Warnf(format, args...)
}

// Error логирует ошибку
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Errorf логирует форматированную ошибку
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zap.Sugar().Errorf(format, args...)
}

// Fatal логирует ошибку и завершает программу
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// Fatalf логирует форматированную ошибку и завершает программу
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zap.Sugar().Fatalf(format, args...)
}

// Sync флашит буферы логгера
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Глобальные функции для удобного использования

// Info логирует информационное сообщение
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Infof логирует форматированное информационное сообщение
func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

// Warn логирует предупреждение
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Warnf логирует форматированное предупреждение
func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

// Error логирует ошибку
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Errorf логирует форматированную ошибку
func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

// Fatal логирует ошибку и завершает программу
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// Fatalf логирует форматированную ошибку и завершает программу
func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

// Sync флашит буферы логгера
func Sync() error {
	return Get().Sync()
}
