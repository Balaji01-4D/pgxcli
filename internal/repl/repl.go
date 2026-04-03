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
	"github.com/muesli/termenv"
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

var chromaFormatter = detectTerminalColorProfile()

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
	prompt  prompt.Prompter
	client  Client
	config  *config.Config
	logger  *slog.Logger
}

func New(client Client, cfg *config.Config, logger *slog.Logger) (*Repl, error) {
	p, err := prompt.New()
	if err != nil {
		return nil, err
	}

	history, entries := newHistory(cfg.Main.HistoryFile, logger)
	if err := applyPromptOptions(p, cfg, entries); err != nil {
		return nil, err
	}

	repl := &Repl{client: client, config: cfg, logger: logger, prompt: p}
	repl.history = history
	return repl, nil
}

func (r *Repl) Read(prefix string, ctx context.Context) (string, error) {
	r.prompt.SetPrefix(prefix)
	text, err := r.prompt.Prompt(ctx)
	if err != nil {
		return "", err
	}
	return text, nil
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

	for {
		suffixStr := r.client.ParsePrompt(r.config.Main.Prompt)
		query, err := r.Read(suffixStr, ctx)
		if err != nil {
			r.logger.Error("error reading input", "error", err)
			r.PrintError(err)
			continue
		}

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
			tw, err := res.Render()
			if err != nil {
				r.logger.Error("error rendering query result", "error", err)
				r.PrintError(err)
				continue
			}
			output := tw.Render()
			// If columns exist, we printed a table. Append the command tag (e.g., "SELECT 5", "INSERT 0 1").
			// If no columns, we just print the command tag.
			if len(res.Columns()) == 0 {
				output = res.CommandTag()
			} else {
				output += "\n" + res.CommandTag()
			}
			r.PrintViaPager(output)
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

func (r *Repl) Close() {
	r.logger.Debug("REPL closing, saving history")
	r.history.saveHistory(r.prompt.History())
	r.logger.Info("REPL closed")
}

func getSyntaxHighlighting(style string) (prompt.SyntaxHighlighter, error) {
	return prompt.SyntaxHighlighterChroma("PostgreSQL SQL dialect", chromaFormatter, style)
}

func applyPromptOptions(p prompt.Prompter, config *config.Config, histories []prompt.HistoryCommand) error {
	highlighter, err := getSyntaxHighlighting(config.Main.Style)
	if err != nil {
		return err
	}
	p.SetSyntaxHighlighter(highlighter)
	p.SetHistory(histories)
	return nil
}

func detectTerminalColorProfile() string {
	switch termenv.ColorProfile() {
	case termenv.TrueColor:
		return "terminal16m"
	case termenv.ANSI256:
		return "terminal256"
	case termenv.ANSI:
		return "terminal16"
	default:
		return "noop" // Chroma's no-op formatter
	}
}
