package repl

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/balaji01-4d/pgxcli/internal/config"
)

const maxHistoryLines = 1000

type history struct {
	entries   []string
	loadCount int
}

func newHistory(historyPath string) *history {
	h := &history{}
	if historyPath == "" || historyPath == config.Default {
		h.loadHistory(getHistoryFilePath())
	} else {
		h.loadHistory(historyPath)
	}
	return h
}

func (h *history) loadHistory(path string) {
	history, err := loadHistoryFromFile(path, maxHistoryLines)
	if err != nil {
		h.entries = []string{}
		h.loadCount = 0
		return
	}
	h.entries = history
	h.loadCount = len(history)
}

func (h *history) saveHistory() {
	historyFilePath := getHistoryFilePath()

	// Only save new commands added during this session
	newCommands := h.entries[h.loadCount:]
	if len(newCommands) == 0 {
		return
	}

	f, err := os.OpenFile(historyFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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

func loadHistoryFromFile(filePath string, maxHistoryLines int) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var history []string

	scanner := bufio.NewScanner(f)
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
