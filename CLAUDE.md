# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build to bin/app
make test           # go test ./...
make lint           # golangci-lint run
make precommit      # lint + test (run before committing)
make clean          # remove bin/

go test ./internal/config      # test a single package
go test -v ./internal/database # verbose
```

## Architecture

`pgxcli` is an interactive PostgreSQL REPL CLI in Go. It follows a strict layered design:

```
cmd/pgxcli/main.go        → entry point; wires context, printer, cobra root command
internal/cli/             → cobra command, flags, password prompting
internal/app/             → REPL loop, command routing, history, rendering
internal/database/        → pgx connection, query/exec dispatch, special commands
internal/parser/          → SQL splitting and query-vs-exec classification
internal/config/          → TOML config load/validate (embeds defaults)
internal/completer/       → keyword autocompletion with schema metadata
internal/cliio/           → Printer interface (stdout/stderr abstraction)
internal/ui/              → Charm TUI forms (bubbletea/huh/lipgloss)
internal/logger/          → slog-based file logger
```

### REPL data flow

1. `app.pgxCLI.Start()` reads input via `go-prompter`-backed `Reader`
2. Input is matched: builtin (e.g. clear screen) → special pgSQL command (`\d`, `\q`, `\c`) → SQL
3. `parser.SplitSqlStatement` splits multi-statement input; everything is routed to `Query` on pgx
4. `database.executor` runs the statement and returns a `result.QueryResult`
5. `app/renderer` formats the result as a table (via `olekukonko/tablewriter`); `Printer` outputs via pager

### Special commands

`\d`, `\l`, `\dt`, `\q`, `\c`, `\conninfo` etc. are handled by the external `pgxspecial` package (`github.com/balaji01-4d/pgxspecial`). `database.executor.executeSpecial` wraps it; rows are materialized into `result.SpecialRow` before being handed back to the REPL.

### Configuration

- Config file auto-created at `~/.config/pgxcli/config.toml` on first run
- `"default"` is a sentinel value; `config.Load()` resolves it to OS-appropriate paths
- `config.validate()` blocks startup on invalid values
- `OnErrorAction`: `STOP` aborts multi-statement execution on error; `RESUME` continues

### Prompt placeholders

`\u` user, `\d` database, `\h` host (short), `\H` host (full), `\p` port, `\t` timestamp, `\n` newline — resolved in `database.Client.ParsePrompt`.

## Conventions

**Errors:** wrap with `fmt.Errorf("context: %w", err)`; log with key `"error"` (not `"err"`): `logger.Error("msg", "error", err)`.

**Logging:** `internal/logger` initializes a file-backed `slog.Logger`; pass the logger down; never use `log` package directly.

**Testing:** use `testify`; mock `conn` interface via `MockConn`/`MockRows` in `database/mocks_test.go`; config tests use `t.TempDir()` with `HOME`/`XDG_CONFIG_HOME` overridden.

**Lint:** golangci-lint is configured in `.golangci.yml` with `revive`, `misspell`, `gocyclo` (min 15), `goconst`, `unconvert`, `unparam`. Tests are excluded from linting. `internal/completer` is excluded from lint paths.

**Interfaces:** `Application`, `Reader`, `Printer`, `Connector` — prefer programming to interfaces for testability.
