# pgxcli

pgxcli is a modern, interactive command-line client for PostgreSQL, written in Go. It aims to provide a rich user experience with features like autocompletion, syntax highlighting, and smart command history.

> **NOTE:** `pgxcli` is currently in active development. The upcoming first release will feature basic syntax highlighting and word-based suggestions. Advanced features like context-aware autocompletion and further quality improvements are planned for future releases.

## Roadmap

- [x] Basic Syntax Highlighting
- [x] Word-based Suggestions
- [ ] Context-aware Autocompletion (Planned)
- [ ] Quality Improvements

## Features

*   **Interactive REPL**: A powerful Read-Eval-Print Loop with a customizable prompt.
*   **Autocompletion**: Currently supports word-based suggestions. Context-aware suggestions are planned for future releases.
*   **Syntax Highlighting**: Colorful output for SQL queries and results.
*   **Special Commands**: Support for standard PostgreSQL backslash commands (e.g., `\d`, `\l`).
*   **Smart History**: Persists your command history.

## Installation

### From Source

Ensure you have Go installed (version 1.21+ recommended).

```bash
git clone https://github.com/balaji01-4d/pgxcli.git
cd pgxcli
make build
```

The binary will be created in `bin/app` (or just `pgxcli` if you adjust the build).

## Usage

Basic usage to connect to a database:

```bash
./bin/app [DBNAME] [USERNAME] [flags]
```

### Examples

Connect to a database named `mydb` as user `myuser`:

```bash
./bin/app mydb myuser
```

Connect using flags:

```bash
./bin/app --host localhost --port 5432 --user postgres --dbname postgres
```

Connect using a connection URI:

```bash
./bin/app postgres://user:password@localhost:5432/dbname
```

### Flags

| Flag | Shorthand | Description |
| :--- | :--- | :--- |
| `--host` | `-h` | Host address of the Postgres database (default "localhost") |
| `--port` | `-p` | Port number (default 5432) |
| `--user` | `-u`, `-U` | Username to connect as |
| `--dbname` | `-d` | Database name to connect to |
| `--password` | `-W` | Force password prompt |
| `--no-password` | `-w` | Never prompt for password |
| `--debug` | | Enable debug mode for verbose logging |
| `--help` | | Show help message |

## Project Structure

The project follows a standard Go layout:

```text
pgxcli/
├── cmd/
│   └── pgxcli/          # Application entry point
├── internal/
│   ├── cli/             # Cobra command definitions and CLI logic
│   ├── completer/       # SQL autocompletion engine and metadata handling
│   ├── config/          # Configuration management
│   ├── database/        # PostgreSQL connection and execution wrapper (using pgx)
│   ├── logger/          # Logging utilities
│   ├── parser/          # SQL parsing (using pg_query_go) for classification
│   └── repl/            # The Read-Eval-Print-Loop core
│       ├── commands/    # Built-in REPL commands (e.g., clear)
│       └── renderer/    # Output rendering and formatting
├── bin/                 # Compiled binaries
├── go.mod               # Go module definition
├── Makefile             # Build automation
└── README.md            # Project documentation
```

## Internal Architecture

*   **REPL**: Built using `elk-language/go-prompt`, it handles user input, history, and rendering.
*   **Database**: Uses `jackc/pgx` for high-performance PostgreSQL interaction. The `internal/database` package abstracts connection pooling and query execution.
*   **Parser**: Integrates `pganalyze/pg_query_go` to analyze SQL statements, determining if they are queries or execution commands to handle transactions and results correctly.
*   **Completer**: Maintains metadata about the database schema (tables, columns) to provide intelligent suggestions.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
