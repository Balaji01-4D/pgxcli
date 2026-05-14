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

	tea "charm.land/bubbletea/v2"
	"github.com/balaji01-4d/pgxcli/internal/app/commands"
	"github.com/balaji01-4d/pgxcli/internal/app/renderer"
	"github.com/balaji01-4d/pgxcli/internal/app/ui"
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
	Start(ctx context.Context, client *database.Client) error

	SetAutoCompleteKeywords(keywords []string)

	// Close performs saving history before exiting.
	Close() error
}

var builtinsCommand = map[string]func(){
	"\\clear": commands.ClearScreen,
}

// pgxCLI is the main implementation of the Application interface.
type pgxCLI struct {
	model   *ui.Model
	program *tea.Program
	Printer cliio.Printer
	config  *config.Config
	logger  *slog.Logger

	pgkws []string
}

func New(cfg *config.Config, printer cliio.Printer, logger *slog.Logger) (Application, error) {
	return &pgxCLI{
		config:  cfg,
		logger:  logger,
		Printer: printer,
	}, nil
}

func (p *pgxCLI) SetAutoCompleteKeywords(keywords []string) {
	p.pgkws = keywords
}

func (p *pgxCLI) execute(ctx context.Context, client *database.Client, query string) tea.Cmd {
	promptReady := func() tea.Msg {
		prefix := client.ParsePrompt(p.config.Main.Prompt)
		return ui.ReadyMsg{Prefix: prefix} // this is used to unblock input after executing a command
	}

	p.logger.Debug("received command", "command_length", len(query))

	if cmd, ok := builtinsCommand[query]; ok {
		p.logger.Debug("executing builtin command", "command", query)
		cmd()
		return promptReady
	}

	return func() tea.Msg {
		metaResult, okay, err := client.ExecuteSpecial(ctx, query)
		if err != nil {
			p.logger.Error("error executing special command", "error", err)
			return ui.ExecCmdMsg{Cmd: tea.Sequence(p.printError(err), promptReady)}
		}
		if okay {
			start := time.Now()
			p.logger.Debug("special command executed", "result_kind", metaResult.ResultKind())
			result, quit, err := p.handleSpecialCommand(ctx, metaResult, client)
			if quit {
				p.logger.Info("REPL exiting via quit command")
				return ui.ExecCmdMsg{Cmd: tea.Quit}
			}

			if err != nil {
				p.logger.Error("error handling special command", "error", err)
				errCmd := p.printError(err)
				return ui.ExecCmdMsg{Cmd: tea.Sequence(errCmd, promptReady)}
			}
			execTime := time.Since(start)
			timingInfo := fmt.Sprintf("Time %.3fs", execTime.Seconds())
			return ui.ExecCmdMsg{Cmd: tea.Sequence(
				p.printViaPager(result+timingInfo),
				promptReady,
			)}
		}

		p.logger.Debug("executing query")
		stmts := parser.SplitSQLStatements(query)
		cmds := make([]tea.Cmd, 0, len(stmts)+1) // +1 for prompt ready

	StatementsLoop:
		for _, stmt := range stmts {
			p.logger.Debug("parsed statement", "statement", stmt)
			if stmt == "" || stmt == ";" {
				continue
			}

			queryResult, err := client.ExecuteQuery(ctx, stmt)
			if err != nil {
				p.logger.Error("query execution failed", "error", err)
				cmds = append(cmds, p.printError(err))
				if p.config.Main.OnError == config.OnErrorStop {
					break StatementsLoop
				}
				continue
			}
			resultCmd, err := p.handleQueryResult(queryResult)
			if err != nil {
				p.logger.Error("error handling query result", "error", err)
				cmds = append(cmds, p.printError(err))
				if p.config.Main.OnError == config.OnErrorStop {
					break StatementsLoop
				}
				continue
			}
			cmds = append(cmds, resultCmd)
		}
		cmds = append(cmds, promptReady)

		return ui.ExecCmdMsg{Cmd: tea.Sequence(cmds...)}
	}
}

func (p *pgxCLI) Start(ctx context.Context, client *database.Client) error {
	executeFunc := func(query string) tea.Cmd {
		return p.execute(ctx, client, query)
	}

	initialPrefix := client.ParsePrompt(p.config.Main.Prompt)
	m, err := ui.New(initialPrefix, p.pgkws, p.config.Main.HistoryFile, executeFunc)
	if err != nil {
		return fmt.Errorf("creating UI model: %w", err)
	}

	p.model = m
	p.program = tea.NewProgram(p.model, tea.WithContext(ctx))

	if _, err := p.program.Run(); err != nil {
		return fmt.Errorf("running UI program: %w", err)
	}

	return nil
}

func (p *pgxCLI) handleSpecialCommand(ctx context.Context, metaResult pgxspecial.SpecialCommandResult, client *database.Client) (string, bool, error) {
	switch metaResult.ResultKind() {

	case database.Exit:
		return "", true, nil

	case database.ChangeDB:
		s := metaResult.(database.ChangeDbAction).Name
		if s != "" {
			if err := client.ChangeDatabase(ctx, s); err != nil {
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

		return fmt.Sprintf(
			"You are connected to database %q as user %q on %s at port %s",
			client.GetDatabase(), client.GetUser(), host, port,
		), false, nil

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

func (p *pgxCLI) handleQueryResult(r result.Result) (tea.Cmd, error) {
	res, ok := r.(*result.QueryResult)
	if !ok {
		return nil, fmt.Errorf("unsupported query result type: %T", r)
	}

	var s strings.Builder
	if err := renderer.Table(res, &s, p.config); err != nil {
		return nil, err
	}

	output := s.String()
	if len(res.Columns()) == 0 {
		output = res.CommandTag()
	} else {
		output += res.CommandTag()
	}

	// Append timing info to the output
	timingInfo := fmt.Sprintf("\nTime %.3fs", res.Duration().Seconds())
	output += timingInfo

	return p.printViaPager(output), nil
}

func (p *pgxCLI) printViaPager(str string) tea.Cmd {
	if p.Printer.ShouldUsePager(str) {
		cmd, ok := cliio.PagerCmd(str)
		if !ok {
			return ui.PrintCmd(str)
		}
		return ui.ShowPagerCmd(cmd)
	}

	return ui.PrintCmd(str)
}

func (p *pgxCLI) printError(err error) tea.Cmd {
	return ui.PrintErrCmd(err)
}

func (p *pgxCLI) Close() error {
	p.logger.Info("closing application and saving history")
	if p.model != nil {
		return p.model.Close()
	}
	return nil
}
