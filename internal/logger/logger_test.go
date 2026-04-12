package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger_CreatesLogFileWithOwnerOnlyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission mode assertion is not reliable on Windows")
	}

	logPath := filepath.Join(t.TempDir(), "app.log")

	logger, err := InitLogger(false, logPath)
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = logger.Close()
	})

	logger.Info("test log entry")

	info, err := os.Stat(logPath)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestInitLogger_ReturnsErrorWhenPathIsDirectory(t *testing.T) {
	dirPath := t.TempDir()

	logger, err := InitLogger(false, dirPath)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestNopLogger(t *testing.T) {
	logger := NopLogger()
	assert.Nil(t, logger.file)
	assert.NotNil(t, logger.Logger)

	logger.Info("this should be discarded")
}

func TestClose_WithNilFile(t *testing.T) {
	logger := &Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		file:   nil,
	}

	err := logger.Close()
	assert.NoError(t, err)
}
