---
title: Getting Started
description: Learn how to install and use pgxcli, a modern PostgreSQL command-line client.
---

## Installation

### From Source

Ensure you have Go installed (version 1.21+ recommended).

```bash
git clone https://github.com/balaji01-4d/pgxcli.git
cd pgxcli
make build
```

The binary will be created in `bin/app`.

## Quick Start

Connect to a database named `mydb` as user `myuser`:

```bash
./bin/app mydb myuser
```

## Connection Options

### Using Flags

```bash
./bin/app --host localhost --port 5432 --user postgres --dbname postgres
```

### Using a Connection URI

```bash
./bin/app postgres://user:password@localhost:5432/dbname
```

## Interactive Mode

Once connected, you can:

- Execute SQL queries with syntax highlighting
- Use tab completion for suggestions
- Access command history with arrow keys
- Run PostgreSQL special commands (e.g., `\d`, `\l`)

## Next Steps

- Check out the [CLI reference](/reference/cli-reference/) for all available flags and commands
- Learn about configuration options and customization
