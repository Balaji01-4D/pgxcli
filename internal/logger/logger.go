package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger wraps slog.Logger and the underlying file for proper cleanup.
type Logger struct {
	*slog.Logger
	file *os.File
}

// InitLogger creates a new structured logger with the specified debug level.
// It writes to a file (creating parent directories if needed) and returns
// a Logger wrapper for proper resource management.
func InitLogger(debug bool, filename string) *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if debug {
		opts.Level = slog.LevelDebug
	}

	// Ensure parent directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Fall back to stderr if we can't create directory
		fmt.Fprintf(os.Stderr, "warning: could not create log directory %s: %v\n", dir, err)
		return newStderrLogger(opts)
	}

	file, err := os.OpenFile(filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Fall back to stderr if we can't open file
		fmt.Fprintf(os.Stderr, "warning: could not open log file %s: %v\n", filename, err)
		return newStderrLogger(opts)
	}

	handler := slog.NewTextHandler(file, opts)
	return &Logger{
		Logger: slog.New(handler),
		file:   file,
	}
}

// newStderrLogger creates a logger that writes to stderr as fallback.
func newStderrLogger(opts *slog.HandlerOptions) *Logger {
	handler := slog.NewTextHandler(os.Stderr, opts)
	return &Logger{
		Logger: slog.New(handler),
		file:   nil,
	}
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
