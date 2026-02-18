package repl

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/repl/commands"
	render "github.com/balaji01-4d/pgxcli/internal/repl/renderer"
	"github.com/balaji01-4d/pgxspecial"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-prompter/prompt"
	"golang.org/x/term"
)

var (
	printErr  = color.New(color.FgHiRed).FprintfFunc()
	printInfo = color.New(color.FgWhite).FprintfFunc()
	printTime = color.New(color.FgHiCyan).FprintfFunc()
)

const (
	DefaultPrompt = `\u@\h:\d> `
	MaxLenPrompt  = 30
)

const (
	language = "PostgreSQL SQL dialect"
	formatter = "terminal256"
)

var builtinsCommand = map[string] func () {
	"clear": commands.ClearScreen,
}

type Client interface {
	GetUser() string
	GetDatabase() string
	GetHost() string
	GetPort() uint16

	ChangeDatabase(ctx context.Context, name string) error
	ParsePrompt(promptStr string) string

	ExecuteQuery(ctx context.Context, query string) (database.Result, error)
	ExecuteSpecial(ctx context.Context, query string) (pgxspecial.SpecialCommandResult, bool, error)
}

type Repl struct {
	history *history
	client  Client
	config  *config.Config
	logger *slog.Logger
	prompt prompt.Prompter
}

func New(client Client, cfg *config.Config, logger *slog.Logger) (*Repl, error) {
	p, err := prompt.New()
	if err != nil {
		return nil, err
	}

	repl := &Repl{client: client, config: cfg, logger: logger, prompt: p}
	repl.history = newHistory(cfg.Main.HistoryFile)
	return repl, nil
}

func (r *Repl) Read(ctx context.Context, prefix string) string {
	r.prompt.SetPrefix(prefix)
	text, _ := r.prompt.Prompt(ctx)
	r.history.append(text)
	return text
}

func (r *Repl) ReadPassword(user string) (string, error) {
	r.Print(fmt.Sprintf("Password for %s: ", user))
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()

	return string(pass), nil
}

func (r *Repl) PrintError(err error) {
	printErr(os.Stderr, "%v\n", err)
}

func (r *Repl) PrintTime(time time.Duration) {
	printTime(os.Stderr, "Time: %.3fs\n", time.Seconds())
}

func (r *Repl) Print(str string) {
	printInfo(os.Stdout, str)
}

func (r *Repl) PrintViaPager(output string) {
	EchoViaPager(func(w io.Writer) error {
		io.WriteString(w, output)
		return nil
	})
}

func (r *Repl) Run(ctx context.Context) {
	err := SetSyntaxHighlighter(r.prompt, r.config.Main.Style)
	if err != nil {
		r.PrintError(err)
		return 
	}
	setAutoCompleter(r.prompt)

	r.history.loadHistory()

	for {
		suffixStr := r.client.ParsePrompt(r.config.Main.Prompt)
		query := r.Read(ctx, suffixStr)

		if strings.TrimSpace(query) == "" {
			continue
		}
		start := time.Now()


		if cmd, ok := builtinsCommand[query]; ok {
			cmd()
			continue
		}
		
		metaResult, okay, err := r.client.ExecuteSpecial(ctx, query)
		if err != nil {
			r.logger.Error("Error executing special command", "err", err)
			r.PrintError(err)
			continue
		}
		if okay {
			result, quit, err := r.handleSpecialCommand(ctx, metaResult)
			if quit {
				return
			}

			if err != nil {
				r.PrintError(err)
				continue
			}
			execTime := time.Since(start)
			r.PrintViaPager(result)
			r.PrintTime(execTime)
			continue
		}

		queryResult, err := r.client.ExecuteQuery(ctx, query)
		if err != nil {
			r.PrintError(err)
			continue
		}
		switch res := queryResult.(type) {
		case *database.QueryResult:
			tw, err := res.Render()
			if err != nil {
				r.PrintError(err)
				continue
			}
			r.Print(tw.Render())
			fmt.Println()
			r.PrintTime(res.Duration())
			continue
		case *database.ExecResult:
			r.Print(res.Status)
			fmt.Println()
			r.PrintTime(res.Duration)
			continue
		}
	}
}

func (r *Repl) handleSpecialCommand(ctx context.Context, metaResult pgxspecial.SpecialCommandResult) (string, bool, error) {

	switch metaResult.ResultKind() {

	case database.Exit:
		return "", true, nil

	case database.ChangeDB:
		s := metaResult.(database.ChangeDbAction).Name
		if s != "" {
			err := r.client.ChangeDatabase(ctx, s)
			if err != nil {
				return "", false, err
			}
		}
		return fmt.Sprintf(
			"You are now connected to database %q as user %q",
			r.client.GetDatabase(),
			r.client.GetUser(),
		), false, nil

	case database.Conninfo:
		var host string
		if strings.HasPrefix(r.client.GetHost(), "/") {
			host = fmt.Sprintf("Socket %q", r.client.GetHost())
		} else {
			host = fmt.Sprintf("Host %q", r.client.GetHost())
		}

		var port string
		if r.client.GetPort() == 0 {
			port = "None"
		} else {
			port = strconv.Itoa(int(r.client.GetPort()))
		}

		info := fmt.Sprintf(
			"You are connected to database %q as user %q on %s at port %s",
			r.client.GetDatabase(), r.client.GetUser(), host, port,
		)
		return info, false, nil

	case pgxspecial.ResultKindRows:
		table, err := render.RenderRowsResult(metaResult)
		if err != nil {
			return "", false, err
		}
		return table.Render(), false, nil

	case pgxspecial.ResultKindDescribeTable:
		tables, err := render.RenderDescribeTableResult(metaResult)
		if err != nil {
			r.logger.Error("Error rendering describe table result", "err", err)
			return "", false, err
		}
		return render.RenderTables(tables, table.StyleBold), false, nil

	case pgxspecial.ResultKindExtensionVerbose:
		tables, err := render.RenderExtensionVerboseResult(metaResult)
		if err != nil {
			return "", false, err
		}
		return render.RenderTables(tables, table.StyleBold), false, nil

	default:
		return "", false, nil
	}
}

// func (r *Repl) getPromptOptions(prefix string) []prompt.Option {
// 	return []prompt.Option{
// 		prompt.WithPrefix(prefix),
// 		prompt.WithHistory(r.history.entries),
// 		prompt.WithTitle("pgxcli"),
// 		prompt.WithHistorySize(100),
// 	}
// }


func (r *Repl) Close() {
	r.history.saveHistory()
}

func SetSyntaxHighlighter(p prompt.Prompter, theme string) error {
	h, err := prompt.SyntaxHighlighterChroma(language, formatter, theme)
	if err != nil {
		return err
	}
	p.SetSyntaxHighlighter(h)
	return nil
}

func setAutoCompleter(p prompt.Prompter) {
	p.SetAutoCompleter(prompt.AutoCompleteSQLKeywords())
}