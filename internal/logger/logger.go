package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

// Logger wraps slog.Logger and the underlying file for proper cleanup.
type Logger struct {
	*slog.Logger
	file *os.File
}

// InitLogger creates a new structured logger with the specified debug level.
// It writes to a file (creating parent directories if needed) and returns
// a Logger wrapper for proper resource management.
func InitLogger(debug bool, filename string) (*Logger, error) {
	if filename == "" || filename == "default" {
		var err error
		filename, err = getDefaultLogPath()
		if err != nil {
			return nil, err
		}
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if debug {
		opts.Level = slog.LevelDebug
	}

	file, err := os.OpenFile(filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	handler := slog.NewTextHandler(file, opts)
	return &Logger{
		Logger: slog.New(handler),
		file:   file,
	}, nil
}

// Close closes the underlying log file if one exists.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// NopLogger returns a logger that discards all output.
// Useful for testing or when logging is disabled.
func NopLogger() *Logger {
	handler := slog.NewTextHandler(io.Discard, nil)
	return &Logger{
		Logger: slog.New(handler),
		file:   nil,
	}
}

func getDefaultLogPath() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
	case "darwin":
		baseDir = filepath.Join(os.Getenv("HOME"), "Library", "Logs")
	default: // Linux and others
		if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
			baseDir = xdg
		} else {
			baseDir = filepath.Join(os.Getenv("HOME"), ".local", "state")
		}
	}

	logDir := filepath.Join(baseDir, "pgxcli")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}

	return filepath.Join(logDir, "pgxcli.log"), nil
}
