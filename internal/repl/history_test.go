package repl

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jedib0t/go-prompter/prompt"
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
		{name: "with empty history file uses default path", path: "", expectedPath: getHistoryFilePath()},
		{name: "with custom history file", path: "/custom_path", expectedPath: "/custom_path"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h, _ := newHistory(test.path, logger)
			assert.Equal(t, test.expectedPath, h.path)
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

	h := history{path: tempFile.Name(), loadCount: 1, logger: testLogger()}
	entries := []prompt.HistoryCommand{
		{Command: "select 1"},
		{Command: "select 2"},
		{Command: "select 3"},
	}
	h.saveHistory(entries)

	data, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 2)

	var got []prompt.HistoryCommand
	for _, line := range lines {
		var entry prompt.HistoryCommand
		err := json.Unmarshal([]byte(line), &entry)
		assert.NoError(t, err)
		got = append(got, entry)
	}

	assert.Equal(t, entries[1:], got)
}

func TestHistorySaveHistory_EntriesShorterThanLoadCount(t *testing.T) {
	tempFile, err := os.CreateTemp("", "history_test")
	assert.NoError(t, err)

	defer func() {
		closeErr := tempFile.Close()
		assert.NoError(t, closeErr)
		err := os.Remove(tempFile.Name())
		assert.NoError(t, err)
	}()

	h := history{path: tempFile.Name(), loadCount: 3, logger: testLogger()}
	entries := []prompt.HistoryCommand{
		{Command: "select 1"},
	}

	assert.NotPanics(t, func() {
		h.saveHistory(entries)
	})

	data, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "", string(data))
}

func TestHistorySaveHistory_CreatesMissingParentDirectoryAndFile(t *testing.T) {
	baseDir, err := os.MkdirTemp("", "history_parent_test")
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(baseDir)
		assert.NoError(t, err)
	}()

	historyPath := filepath.Join(baseDir, "non_exist_folder", "user_provided_name.jsonl")
	h := history{path: historyPath, loadCount: 0, logger: testLogger()}
	entries := []prompt.HistoryCommand{{Command: "select 1"}}

	h.saveHistory(entries)

	_, err = os.Stat(filepath.Dir(historyPath))
	assert.NoError(t, err)

	data, err := os.ReadFile(historyPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, strings.TrimSpace(string(data)))
}

func TestHistorySaveHistory_DisablesHistoryOnInvalidPath(t *testing.T) {
	h := history{path: "\x00", loadCount: 0, logger: testLogger()}
	entries := []prompt.HistoryCommand{{Command: "select 1"}}

	h.saveHistory(entries)

	assert.True(t, h.disabled)

	// Second call should no-op because history is disabled.
	assert.NotPanics(t, func() {
		h.saveHistory(entries)
	})
}

func TestLoadHistory(t *testing.T) {
	r := strings.NewReader(strings.Join([]string{
		`{"command":"query1","timestamp":"2026-04-04T10:00:00Z"}`,
		`{"command":"query2","timestamp":"2026-04-04T10:00:01Z"}`,
		"not-valid-json",
		`{"command":"query3","timestamp":"2026-04-04T10:00:02Z"}`,
		`{"command":"query4","timestamp":"2026-04-04T10:00:03Z"}`,
	}, "\n"))

	actual, err := loadHistory(r, 3, testLogger())
	assert.NoError(t, err)
	commands := make([]string, 0, len(actual))
	for _, entry := range actual {
		commands = append(commands, entry.Command)
	}
	assert.Equal(t, []string{"query2", "query3", "query4"}, commands)
}
