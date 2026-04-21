package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	render "github.com/balaji01-4d/pgxcli/internal/app/renderer"
	"github.com/balaji01-4d/pgxcli/internal/cliio"
	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/parser"
	"github.com/balaji01-4d/pgxspecial"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Application interface {
	Start(ctx context.Context, client *database.Client)
	Close() error
}

type PgxCLI struct {
	prompt  Reader
	Printer cliio.Printer
	History *history
	config  *config.Config
	logger  *slog.Logger
}

func NewPgxCLI(config *config.Config, printer cliio.Printer, logger *slog.Logger) (*PgxCLI, error) {
	history, entries := newHistory(config.Main.HistoryFile, logger)
	reader, err := NewPgxReader()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt reader: %w", err)
	}
	applyReaderOptions(reader, config, entries)

	return &PgxCLI{
		config: config, logger: logger, prompt: reader, Printer: printer, History: history,
	}, nil
}

func (p *PgxCLI) Start(ctx context.Context, client *database.Client) {
	for {
		suffixStr := client.ParsePrompt(p.config.Main.Prompt)
		rawInput, err := p.prompt.Read(suffixStr, ctx)
		if err != nil {
			p.logger.Error("error reading input", "error", err)
			p.Printer.PrintError(err)
			continue
		}

		trimmedInput := strings.TrimSpace(rawInput)

		if trimmedInput == "" {
			continue
		}
		p.logger.Debug("received command", "command_length", len(trimmedInput))

		if cmd, ok := builtinsCommand[trimmedInput]; ok {
			p.logger.Debug("executing builtin command", "command", trimmedInput)
			cmd()
			continue
		}

		metaResult, okay, err := client.ExecuteSpecial(ctx, trimmedInput)
		if err != nil {
			p.logger.Error("error executing special command", "error", err)
			p.Printer.PrintError(err)
			continue
		}
		if okay {
			start := time.Now()
			p.logger.Debug("special command executed", "result_kind", metaResult.ResultKind())
			result, quit, err := p.handleSpecialCommand(ctx, metaResult, client)
			if quit {
				p.logger.Info("REPL exiting via quit command")
				return
			}

			if err != nil {
				p.logger.Error("error handling special command", "error", err)
				p.Printer.PrintError(err)
				continue
			}
			execTime := time.Since(start)
			p.Printer.PrintViaPager(result)
			p.Printer.PrintTime(execTime)
			continue
		}

		p.logger.Debug("executing query")
		stmts, err := parser.SplitSqlStatement(trimmedInput)
		if err != nil {
			p.logger.Error("failed to split sql statements", "error", err)
			p.Printer.PrintError(err)
			continue
		}

	StatementsLoop:
		for _, stmt := range stmts {
			p.logger.Debug("parsed statement", "statement", stmt)
			if stmt == "" {
				continue
			}

			queryResult, err := client.ExecuteQuery(ctx, stmt)
			if err != nil {
				p.logger.Error("query execution failed", "error", err)
				p.Printer.PrintError(err)
				if p.config.Main.OnError == config.OnErrorStop {
					break StatementsLoop
				}
				continue
			}

			if err := p.handleQueryResult(queryResult); err != nil {
				p.logger.Error("error handling query result", "error", err)
				p.Printer.PrintError(err)
				continue
			}
		}
	}
}

func (p *PgxCLI) Close() error {
	p.History.saveHistory(p.prompt.History())
	return nil
}

func (p *PgxCLI) handleSpecialCommand(ctx context.Context, metaResult pgxspecial.SpecialCommandResult, client *database.Client) (string, bool, error) {
	switch metaResult.ResultKind() {

	case database.Exit:
		return "", true, nil

	case database.ChangeDB:
		s := metaResult.(database.ChangeDbAction).Name
		if s != "" {
			err := client.ChangeDatabase(ctx, s)
			if err != nil {
				return "", false, err
			}
		}
		return fmt.Sprintf(
			"You are now connected to database %q as user %q",
			client.GetDatabase(),
			client.GetUser(),
		), false, nil

	case database.Conninfo:
		var host string
		if strings.HasPrefix(client.GetHost(), "/") {
			host = fmt.Sprintf("Socket %q", client.GetHost())
		} else {
			host = fmt.Sprintf("Host %q", client.GetHost())
		}

		var port string
		if client.GetPort() == 0 {
			port = "None"
		} else {
			port = strconv.Itoa(int(client.GetPort()))
		}

		info := fmt.Sprintf(
			"You are connected to database %q as user %q on %s at port %s",
			client.GetDatabase(), client.GetUser(), host, port,
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
			p.logger.Error("error rendering describe table result", "error", err)
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

func (p *PgxCLI) handleQueryResult(result database.Result) error {
	switch res := result.(type) {
	case *database.QueryResult:
		tw, err := res.Render()
		if err != nil {
			return err
		}
		output := tw.Render()
		// If columns exist, we printed a table. Append the command tag (e.g., "SELECT 5", "INSERT 0 1").
		// If no columns, we just print the command tag.
		if len(res.Columns()) == 0 {
			output = res.CommandTag()
		} else {
			output += "\n" + res.CommandTag()
		}
		p.Printer.PrintViaPager(output)
		p.Printer.PrintTime(res.Duration())
		return nil
	case *database.ExecResult:
		p.Printer.PrintViaPager(res.Status)
		fmt.Println()
		p.Printer.PrintTime(res.Duration)
		return nil
	default:
		return fmt.Errorf("unsupported query result type: %T", result)
	}
}
