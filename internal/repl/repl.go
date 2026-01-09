package repl

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/logger"
	render "github.com/balaji01-4d/pgxcli/internal/repl/renderer"
	"github.com/balaji01-4d/pgxspecial"
	"github.com/elk-language/go-prompt"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
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
	config  config.Config
}

func New(client Client, cfg config.Config) *Repl {
	repl := &Repl{client: client, config: cfg}
	repl.history = newHistory(cfg.Main.HistoryFile)
	return repl
}

func (r *Repl) Read(prefix string) string {
	text := prompt.Input(
		r.getPromptOptions(prefix)...,
	)
	r.history.append(text)
	return text
}

func (r *Repl) ReadPassword() (string, error) {
	r.Print("Enter the password: ")
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
	r.history.loadHistory()

	for {
		suffixStr := r.client.ParsePrompt(r.config.Main.Prompt)
		query := r.Read(suffixStr)

		if strings.TrimSpace(query) == "" {
			continue
		}
		start := time.Now()

		metaResult, okay, err := r.client.ExecuteSpecial(ctx, query)
		if err != nil {
			logger.Log.Error("Error executing special command", "err", err)
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
			logger.Log.Error("Error rendering describe table result", "err", err)
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
