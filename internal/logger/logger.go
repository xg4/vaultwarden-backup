package logger

import (
	"log/slog"
	"os"
)

// Setup 初始化并配置日志记录器
func Setup() {
	logLevel := getLogLevel()
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))

	slog.Info("日志级别已设置", "level", logLevel.String())
}

// getLogLevel 根据环境变量设置日志级别
func getLogLevel() slog.Level {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG", "debug":
		return slog.LevelDebug
	case "INFO", "info":
		return slog.LevelInfo
	case "WARN", "warn", "WARNING", "warning":
		return slog.LevelWarn
	case "ERROR", "error":
		return slog.LevelError
	default:
		// 生产环境默认使用 INFO 级别
		if os.Getenv("ENV") == "production" {
			return slog.LevelWarn
		}
		return slog.LevelInfo
	}
}
