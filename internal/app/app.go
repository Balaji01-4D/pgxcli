// Package app is the application layer.
//
// It orchestrates the repl loop, command routing, result handling and history management.
// custom commands and edit command buffer are also handled here.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/app/renderer"
	"github.com/balaji01-4d/pgxcli/internal/cliio"
	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/database/result"
	"github.com/balaji01-4d/pgxcli/internal/parser"
	"github.com/balaji01-4d/pgxspecial"
)

// Application defines the interface for the main application logic.
type Application interface {
	// Start starts the main repl loop, reading input, executing commands and printing results until the user exits.
	Start(ctx context.Context, client *database.Client)

	// SetAutocompleter sets the autocompleter keywords for the prompt.
	SetAutocompleter(keywords []string)

	// Close performs saving history before exiting.
	Close() error
}

// pgxCLI is the main implementation of the Application interface.
// It holds the prompt reader, printer, history manager, configuration and logger.
type pgxCLI struct {
	prompt  Reader
	Printer cliio.Printer
	History *history
	config  *config.Config
	logger  *slog.Logger
}

// New initializes the Application
// based on configuration, it sets up the history manager, and prompt reader.
func New(config *config.Config, printer cliio.Printer, logger *slog.Logger) (Application, error) {
	history, entries := newHistory(config.Main.HistoryFile, logger)
	reader, err := newReader()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt reader: %w", err)
	}
	if err := applyReaderOptions(reader, config, entries); err != nil {
		return nil, fmt.Errorf("failed to apply reader options: %w", err)
	}

	return &pgxCLI{
		config: config, logger: logger, prompt: reader, Printer: printer, History: history,
	}, nil
}

func (p *pgxCLI) Start(ctx context.Context, client *database.Client) {
	for {
		suffixStr := client.ParsePrompt(p.config.Main.Prompt)
		rawInput, err := p.prompt.Read(ctx, suffixStr)
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
		stmts := parser.SplitSQLStatements(trimmedInput)

	StatementsLoop:
		for _, stmt := range stmts {
			p.logger.Debug("parsed statement", "statement", stmt)
			if stmt == "" || stmt == ";" {
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

func (p *pgxCLI) SetAutocompleter(keywords []string) {
	p.prompt.SetAutocompleter(keywords)
}

func (p *pgxCLI) Close() error {
	return p.History.saveHistory(p.prompt.History())
}

func (p *pgxCLI) handleSpecialCommand(ctx context.Context, metaResult pgxspecial.SpecialCommandResult, client *database.Client) (string, bool, error) {
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
		table, err := renderer.RowsResult(metaResult, p.config)
		if err != nil {
			return "", false, err
		}
		return table, false, nil

	case pgxspecial.ResultKindDescribeTable:
		tables, err := renderer.DescribeTableResult(metaResult, p.config)
		if err != nil {
			p.logger.Error("error rendering describe table result", "error", err)
			return "", false, err
		}
		return tables, false, nil

	case pgxspecial.ResultKindExtensionVerbose:
		tables, err := renderer.ExtensionVerboseResult(metaResult, p.config)
		if err != nil {
			return "", false, err
		}
		return tables, false, nil

	default:
		return "", false, nil
	}
}

func (p *pgxCLI) handleQueryResult(r result.Result) error {
	res, ok := r.(*result.QueryResult)
	if !ok {
		return fmt.Errorf("unsupported query result type: %T", r)
	}

	var s strings.Builder
	err := renderer.Table(res, &s, p.config)
	if err != nil {
		return err
	}
	output := s.String()
	// If columns exist, we printed a table. Append the command tag (e.g., "SELECT 5", "INSERT 0 1").
	// If no columns, we just print the command tag.
	if len(res.Columns()) == 0 {
		output = res.CommandTag()
	} else {
		output += res.CommandTag()
	}
	p.Printer.PrintViaPager(output)
	p.Printer.PrintTime(res.Duration())
	return nil
}
