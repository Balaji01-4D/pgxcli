package repl

import (
	"io"
	"os"
	"time"

	"github.com/elk-language/go-prompt"
	"github.com/fatih/color"
)

var (
	printErr  = color.New(color.FgHiRed).FprintfFunc()
	printTime = color.New(color.FgHiCyan).FprintfFunc()
)

type Repl struct {
	history *history
}

func New() *Repl {
	repl := &Repl{}
	repl.history = newHistory(1000, "")
	return repl
}

func (r *Repl) Read(prefix string) string {
	text := prompt.Input(
		r.getPromptOptions(prefix)...,
	)
	r.history.append(text)
	return text
}

func (r *Repl) PrintError(err error) {
	printErr(os.Stderr, "%v\n", err)
}

func (r *Repl) PrintTime(time time.Duration) {
	printTime(os.Stderr, "Time: %.3fs\n", time.Seconds())
}

func (r *Repl) Print(output string) {
	EchoViaPager(func(w io.Writer) error {
		io.WriteString(w, output)
		return nil
	})
}

func (r *Repl) getPromptOptions(prefix string) []prompt.Option {
	return []prompt.Option{
		prompt.WithPrefix(prefix),
		prompt.WithHistory(r.history.entries),
		prompt.WithTitle("pgxcli"),
		prompt.WithHistorySize(100),
	}
}

func (r *Repl) Close() {
	r.history.saveHistory()
}
