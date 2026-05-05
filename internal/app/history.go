package app

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
			h.logger.Error("failed to load history from path", "path", h.path, "error", err)
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
		h.logger.Error("failed to decode history", "path", h.path, "error", err)
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

func (h *history) saveHistory(entries []prompt.HistoryCommand) error {
	if len(entries) <= h.loadCount {
		return nil
	}

	historyDir := filepath.Dir(h.path)
	if err := os.MkdirAll(historyDir, 0o700); err != nil {
		return err
	}

	var commands []prompt.HistoryCommand
	flags := os.O_CREATE | os.O_WRONLY

	if len(entries) > maxHistoryLines {
		commands = entries[max(0, len(entries)-maxHistoryLines):]
		flags |= os.O_TRUNC
	} else {
		commands = entries[h.loadCount:]
		flags |= os.O_APPEND
	}

	f, err := os.OpenFile(h.path, flags, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			h.logger.Error("failed to close history file after writing", "error", err)
		}
	}()

	w := bufio.NewWriter(f)
	for _, entry := range commands {
		line, err := json.Marshal(entry)
		if err != nil {
			h.logger.Warn("skipping entry, failed to marshal", "command", entry.Command, "error", err)
			continue
		}
		if _, err := w.Write(line); err != nil {
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}

	}

	if err := w.Flush(); err != nil {
		return err
	}
	h.logger.Debug("history saved", "entries_written", len(commands))
	return nil
}

func getHistoryFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".pgxcli_history.jsonl")
}
