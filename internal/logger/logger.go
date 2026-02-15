package logger

import (
	"log/slog"
	"os"
)

func InitLogger(debug bool, filename string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if debug {
		opts.Level = slog.LevelDebug
	}

	file, _ := os.OpenFile(filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	handler := slog.NewTextHandler(file, opts)

	return slog.New(handler)
}
