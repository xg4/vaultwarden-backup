package logger

import (
	"log/slog"
	"os"
)

func InitLogger(level slog.Level) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
}

func GetLogger() *slog.Logger {
	return slog.Default()
}
