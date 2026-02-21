package repl

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testLogger returns a logger that discards all output for tests.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}

func TestNewHistory(t *testing.T) {
	logger := testLogger()
	tests := []struct {
		name string
		path string

		expectedPath string
	}{
		{name: "with default history file", expectedPath: getHistoryFilePath()},
		{name: "with custom history file", path: "/custom_path", expectedPath: "/custom_path"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := newHistory(test.path, logger)
			assert.Equal(t, test.expectedPath, actual.path)
		})
	}
}

func TestHistorySaveHistory(t *testing.T) {
	tempFile, err := os.CreateTemp("", "history_test")
	assert.NoError(t, err)

	defer func() {
		closeErr := tempFile.Close()
		assert.NoError(t, closeErr)
		err := os.Remove(tempFile.Name())
		assert.NoError(t, err)
	}()

	histories := []string{"query1", "query2", "query3", "query4"}

	h := history{
		path:   tempFile.Name(),
		logger: testLogger(),
	}
	for _, hist := range histories {
		h.append(hist)
	}
	h.saveHistory()

	data, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)

	expected := strings.Join(histories, "\n") + "\n"
	actual := string(data)

	assert.Equal(t, expected, actual)

}

func TestLoadHistory(t *testing.T) {
	histories := []string{
		"query1",
		"query2",
		"query3",
		"query4",
		"query5",
		"query6",
		"query7",
		"query8",
		"query9",
		"query10",
	}

	r := strings.NewReader(strings.Join(histories, "\n"))
	max := 3

	actual, err := loadHistory(r, max)
	assert.NoError(t, err)
	assert.Equal(t, histories[len(histories)-max:], actual)
}
