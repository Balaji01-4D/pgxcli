package repl

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/jedib0t/go-prompter/prompt"
)

const maxHistoryLines = 1000

type history struct {
	path      string
	loadCount int
	logger    *slog.Logger
}

func newHistory(historyPath string, logger *slog.Logger) (*history, []prompt.HistoryCommand) {
	h := &history{logger: logger}
	if historyPath == "" || historyPath == config.Default {
		h.path = getHistoryFilePath()
	} else {
		h.path = historyPath
	}
	entries := h.loadHistory()
	h.loadCount = len(entries)
	return h, entries
}

func (h *history) loadHistory() []prompt.HistoryCommand {
	file, err := os.Open(h.path)
	if err != nil {
		if !os.IsNotExist(err) {
			h.logger.Warn("could not open history file", "path", h.path, "error", err)
		}
		return []prompt.HistoryCommand{}
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.Error("failed to close history file", "error", err)
		}
	}()

	entries, err := loadHistory(file, maxHistoryLines, h.logger)
	if err != nil {
		h.logger.Error("failed to load history", "error", err)
		return []prompt.HistoryCommand{}
	}
	return entries
}

func loadHistory(r io.Reader, maxHistoryLines int, logger *slog.Logger) ([]prompt.HistoryCommand, error) {
	var entries []prompt.HistoryCommand
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry prompt.HistoryCommand
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			logger.Warn("skipping corrupt history entry", "line", line, "error", err)
			continue
		}
		entries = append(entries, entry)
		if len(entries) > maxHistoryLines {
			entries = entries[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (h *history) saveHistory(entries []prompt.HistoryCommand) {
	newCommands := entries[h.loadCount:]
	if len(newCommands) == 0 {
		return
	}

	f, err := os.OpenFile(h.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		h.logger.Error("failed to open history file for writing", "error", err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			h.logger.Error("failed to close history file after writing", "error", err)
		}
	}()

	w := bufio.NewWriter(f)
	for _, entry := range newCommands {
		line, err := json.Marshal(entry)
		if err != nil {
			h.logger.Warn("skipping entry, failed to marshal", "command", entry.Command, "error", err)
			continue
		}
		w.Write(line)
		w.WriteByte('\n')
	}

	if err := w.Flush(); err != nil {
		h.logger.Error("failed to flush history file", "error", err)
		return
	}
	h.logger.Debug("history saved", "new_entries", len(newCommands))
}

func getHistoryFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".pgxcli_history.jsonl")
}
