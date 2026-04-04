package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger_CreatesLogFileWithOwnerOnlyPermissions(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "app.log")

	logger := InitLogger(false, logPath)
	t.Cleanup(func() {
		_ = logger.Close()
	})

	logger.Info("test log entry")

	info, err := os.Stat(logPath)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestInitLogger_FallsBackToStderrWhenPathIsDirectory(t *testing.T) {
	dirPath := t.TempDir()

	originalStderr := os.Stderr
	reader, writer, err := os.Pipe()
	assert.NoError(t, err)
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = originalStderr
		_ = reader.Close()
	})

	logger := InitLogger(false, dirPath)
	t.Cleanup(func() {
		_ = logger.Close()
	})

	assert.Nil(t, logger.file)

	const logMessage = "stderr fallback test message"
	logger.Info(logMessage)

	_ = writer.Close()
	output, readErr := io.ReadAll(reader)
	assert.NoError(t, readErr)
	assert.Contains(t, string(output), logMessage)
}

func TestNopLogger(t *testing.T) {
	logger := NopLogger()
	assert.Nil(t, logger.file)
	assert.NotNil(t, logger.Logger)

	logger.Info("this should be discarded")
}

func TestNewStderrLogger(t *testing.T) {
	logger := newStderrLogger(&slog.HandlerOptions{})
	assert.Nil(t, logger.file)
	assert.NotNil(t, logger.Logger)
}

func TestClose_WithNilFile(t *testing.T) {
	logger := &Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		file:   nil,
	}

	err := logger.Close()
	assert.NoError(t, err)
}
