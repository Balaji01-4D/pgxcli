# Copilot Instructions for pgxcli

## Build, Test, and Lint Commands

### Build
```bash
make build          # Build the application (output: bin/app)
make clean          # Remove compiled binaries
```

### Test
```bash
go test ./...                      # Run all tests
go test ./internal/config          # Test specific package
go test -v ./internal/database     # Verbose output
```

### Lint
```bash
make lint           # Run golangci-lint (requires golangci-lint binary)
make precommit      # Run lint AND test (equivalent to: lint test)
```

### Run
```bash
make build && ./bin/app mydb myuser          # Build and run
./bin/app --host localhost --user postgres  # With flags
```

## High-Level Architecture

pgxcli is an interactive PostgreSQL CLI written in Go. The application follows a layered architecture:

### Core Packages

- **`cmd/pgxcli/main.go`** – Entry point; initializes context, printer, and root CLI command.
- **`internal/cli`** – Cobra command definitions (flags parsing, password prompting, root command setup).
- **`internal/app`** – Application orchestration layer. `Application` interface manages the REPL loop, routes commands to database or special handlers, renders results, and manages history.
- **`internal/database`** – PostgreSQL abstraction layer using `jackc/pgx`. Contains `Client` (connection mgmt), `Executor` (query/exec routing), and special commands handler.
- **`internal/config`** – Configuration management (loads embedded defaults + user config from `~/.config/pgxcli/config.toml`).
- **`internal/logger`** – Structured logging via `log/slog`. Initializes file-based logger with debug flag support.
- **`internal/completer`** – SQL autocompletion engine; maintains database schema metadata.
- **`internal/parser`** – SQL parsing; classifies statements as queries vs. execution commands via custom splitter.
- **`internal/ui`** – Forms and prompts (uses Charm libraries: bubbletea, huh, lipgloss).
- **`internal/cliio`** – Output printing abstraction (stdout/stderr wrapper).

### Data Flow

1. User enters input → **app.Start()** reads via prompt reader
2. Input routed: builtin command → special pgSQL command (e.g., `\d`) → SQL query
3. **database.Client** determines statement type (query vs. execution) via parser
4. Executor runs query/exec; wraps pgx results
5. **renderer** formats results as tables/text; printer outputs to stdout/stderr
6. History stored in memory and persisted to history file on close

## Key Conventions

### Configuration & Initialization

- Configuration is validated on load (see `config.validate()`). Invalid values block startup.
- Paths support `"default"` as a sentinel value; `config.Load()` resolves it to OS-appropriate defaults (e.g., `~/.config/pgxcli/`).
- Logger accepts `debug bool` flag; enables debug-level slog output if true.

### Testing Patterns

- Tests use **stretchr/testify** for assertions and mocking.
- Database tests define mock doubles (`MockConn`, `MockRows`) in `mocks_test.go`; see `database/executor_test.go` for patterns.
- Config tests use `t.TempDir()` and env isolation (set `HOME`, `XDG_CONFIG_HOME`, etc. to temp dirs).
- Context passed as first argument to all db operations: `ctx` is `context.Background()` or passed from CLI.

### Error Handling & Logging

- All errors are wrapped with context using `fmt.Errorf(..., %w, err)`.
- Logs use structured `slog` with attribute key `"error"` (not `"err"`) for error values: `logger.Error("msg", "error", err)`.
- REPL loop catches errors mid-execution; logs and continues (doesn't exit).

### Special Commands & PostgreSQL Features

- Special pgSQL commands (e.g., `\d`, `\l`) are registered via **pgxspecial** package initialization (see `internal/database/special_commands.go`).
- The `OnErrorAction` config controls multi-statement behavior: `STOP` exits on error, `RESUME` continues.
- **Parser** determines if a statement is a query (SELECT-like) or execution (INSERT/UPDATE/DELETE) to control transaction handling.

### REPL Reader & History

- REPL reader wraps `elk-language/go-prompt` (now uses `jedib0t/go-prompter` in latest version).
- History is in-memory and persisted to a history file (default: `~/.local/share/pgxcli/history` or `~/.pgxcli/history` depending on OS).
- History entries are passed to the reader for autocomplete suggestions.

### Rendering & Output

- Results are rendered as tables by default (via `jedib0t/go-pretty`).
- Renderer lives in `internal/app/renderer`; handles formatting queries, exec results, and errors.
- Output goes through a `Printer` interface (wraps stdout/stderr) for testability.

### Go Version & Dependencies

- Go 1.25.8+ required.
- Key dependencies: pgx/v5 (database), cobra (CLI), viper (config), charm libraries (TUI), testify (testing).

### Naming Conventions

- Unexported (private) functions/vars use `camelCase` (e.g., `pgxCLI`, `newReader()`).
- Exported (public) types/funcs use `PascalCase` (e.g., `Application`, `Client`, `Logger`).
- Interfaces are typically named `Action` or end with `-er` (e.g., `Reader`, `Printer`, `Application`).
- Test files use `_test.go` suffix; test functions start with `Test` (standard Go convention).
