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

var builtinsCommand = map[string]func(){
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
	logger  *slog.Logger
}

func New(client Client, cfg *config.Config, logger *slog.Logger) *Repl {
	repl := &Repl{client: client, config: cfg, logger: logger}
	repl.history = newHistory(cfg.Main.HistoryFile, logger)
	return repl
}

func (r *Repl) Read(prefix string) string {
	text := prompt.Input(
		r.getPromptOptions(prefix)...,
	)
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
	if err := EchoViaPager(func(w io.Writer) error {
		_, err := io.WriteString(w, output)
		return err
	}); err != nil {
		r.logger.Error("failed to write via pager", "error", err)
	}
}

func (r *Repl) Run(ctx context.Context) {
	r.logger.Info("REPL started")
	r.history.loadHistory()

	for {
		suffixStr := r.client.ParsePrompt(r.config.Main.Prompt)
		query := r.Read(suffixStr)

		if strings.TrimSpace(query) == "" {
			continue
		}
		start := time.Now()

		r.logger.Debug("received command", "command_length", len(query))

		if cmd, ok := builtinsCommand[query]; ok {
			r.logger.Debug("executing builtin command", "command", query)
			cmd()
			continue
		}

		metaResult, okay, err := r.client.ExecuteSpecial(ctx, query)
		if err != nil {
			r.logger.Error("error executing special command", "error", err)
			r.PrintError(err)
			continue
		}
		if okay {
			r.logger.Debug("special command executed", "result_kind", metaResult.ResultKind())
			result, quit, err := r.handleSpecialCommand(ctx, metaResult)
			if quit {
				r.logger.Info("REPL exiting via quit command")
				return
			}

			if err != nil {
				r.logger.Error("error handling special command", "error", err)
				r.PrintError(err)
				continue
			}
			execTime := time.Since(start)
			r.PrintViaPager(result)
			r.PrintTime(execTime)
			continue
		}

		r.logger.Debug("executing query")
		queryResult, err := r.client.ExecuteQuery(ctx, query)
		if err != nil {
			r.logger.Error("query execution failed", "error", err)
			r.PrintError(err)
			continue
		}
		switch res := queryResult.(type) {
		case *database.QueryResult:
			if len(res.Columns()) == 0 {
				r.PrintViaPager(res.CommandTag())
				r.PrintTime(res.Duration())
				continue
			}
			if pagerErr := EchoViaPager(func(w io.Writer) error {
				if err := res.RenderTo(w); err != nil {
					return err
				}
				_, err := fmt.Fprintln(w, res.CommandTag())
				return err
			}); pagerErr != nil {
				r.logger.Error("error rendering query result", "error", pagerErr)
				r.PrintError(pagerErr)
			}
			r.PrintTime(res.Duration())
			continue
		case *database.ExecResult:
			r.PrintViaPager(res.Status)
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
			r.logger.Error("error rendering describe table result", "error", err)
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
	r.logger.Debug("REPL closing, saving history")
	r.history.saveHistory()
	r.logger.Info("REPL closed")
}
