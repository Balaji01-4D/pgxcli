package repl

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"github.com/elk-language/go-prompt"
	"github.com/fatih/color"
)

type Repl struct {
	db string
	history []string
}

func NewModel(db string) *Repl {
	repl := &Repl{
		db: db,
	}
	repl.loadHistory()
	return repl
}


func (r *Repl) Read() string {
	text := prompt.Input(
		r.getPromptOptions()...,
	)
	r.addToHistory(text)
	return text
}

func (r *Repl) PrintError(err error) {
	c := color.New(color.FgRed)
	c.Fprintf(os.Stderr, "%v\n", err)
}

func (r *Repl) PrintTime(time time.Duration) {
	c := color.New(color.FgCyan)
	c.Fprintf(os.Stderr, "Time: %.3fs\n", time.Seconds())
}

func (r *Repl) Print(output string) {
	EchoViaPager(func(w io.Writer) error {
		io.WriteString(w, output)
		return nil
	})
}

func (r *Repl) loadHistory() {
	historyFilePath := getHistoryFilePath()
	history, err := loadHistoryFromFile(historyFilePath)
	if err != nil {
		r.history = []string{}
		return
	}
	r.history = history
}

func (r *Repl) saveHistory() {
	historyFilePath := getHistoryFilePath()
	
	f, err := os.OpenFile(historyFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	
	f.WriteString(strings.Join(r.history, "\n"))
}

func (r *Repl) addToHistory(command string) {
	r.history = append(r.history, command)
}

func (r *Repl) getPrefix() string {
	return r.db + "> "
}

func (r *Repl) Close() {
	r.saveHistory()
}

func (r *Repl) getPromptOptions() []prompt.Option {
	return []prompt.Option{
		prompt.WithPrefix(r.getPrefix()),
		prompt.WithHistory(r.history),
		prompt.WithTitle("pgxcli"),
		prompt.WithHistorySize(100), 
	}
}

func getHistoryFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return homeDir + string(os.PathSeparator) + ".pgxcli_history"
}	

func loadHistoryFromFile(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var history []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		history = append(history, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return history, nil
}
