package repl

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/balaji01-4d/pgxcli/internal/config"
)

const maxHistoryLines = 1000

type history struct {
	path      string
	entries   []string
	loadCount int
}

func newHistory(historyPath string) *history {
	h := &history{}
	if historyPath == "" || historyPath == config.Default {
		h.path = getHistoryFilePath()
	} else {
		h.path = historyPath
	}
	return h
}

func (h *history) loadHistory() {
	file, err := os.Open(h.path)
	if err != nil {
		return
	}
	defer file.Close()
	history, err := loadHistory(file, maxHistoryLines)
	if err != nil {
		h.entries = []string{}
		h.loadCount = 0
		return
	}
	h.entries = history
	h.loadCount = len(history)
}

func (h *history) saveHistory() {

	// Only save new commands added during this session
	newCommands := h.entries[h.loadCount:]
	if len(newCommands) == 0 {
		return
	}

	f, err := os.OpenFile(h.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(strings.Join(newCommands, "\n") + "\n")
}

func (h *history) append(command string) {
	h.entries = append(h.entries, command)
}

func getHistoryFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".pgxcli_history")
}

func loadHistory(r io.Reader, maxHistoryLines int) ([]string, error) {
	var history []string

	scanner := bufio.NewScanner(r)
	// Use a circular buffer approach: keep only last N lines
	for scanner.Scan() {
		history = append(history, scanner.Text())
		if len(history) > maxHistoryLines {
			// Remove oldest entry to keep memory bounded
			history = history[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return history, nil
}
