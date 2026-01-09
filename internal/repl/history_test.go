package repl

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHistory(t *testing.T) {
	tests := []struct {
		name string
		path string

		expected *history
	}{
		{name: "with default history file", expected: &history{path: getHistoryFilePath()}},
		{name: "with custom history file", path: "/custom_path", expected: &history{path: "/custom_path"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := newHistory(test.path)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHistorySaveHistory(t *testing.T) {
	tempFile, err := os.CreateTemp("", "history_test")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	histories := []string{"query1", "query2", "query3", "query4"}

	h := history{
		path: tempFile.Name(),
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
