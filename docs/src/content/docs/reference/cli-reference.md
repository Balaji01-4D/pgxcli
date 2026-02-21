---
title: CLI Reference
description: Complete reference for pgxcli command-line flags and options.
---

## Command Syntax

```bash
pgxcli [DBNAME] [USERNAME] [flags]
```

## Connection Flags

### `--host`, `-h`
Host address of the PostgreSQL database.

- **Type**: `string`
- **Default**: `localhost`

### `--port`, `-p`
Port number for the PostgreSQL server.

- **Type**: `integer`
- **Default**: `5432`

### `--user`, `-u`, `-U`
Username to connect as.

- **Type**: `string`
- **Required**: Yes (unless provided as positional argument)

### `--dbname`, `-d`
Database name to connect to.

- **Type**: `string`
- **Required**: Yes (unless provided as positional argument)

### `--password`, `-W`
Force password prompt before connecting.

- **Type**: `boolean`
- **Default**: `false`

### `--no-password`, `-w`
Never prompt for password.

- **Type**: `boolean`
- **Default**: `false`

## Other Options

### `--debug`
Enable debug mode for verbose logging.

- **Type**: `boolean`
- **Default**: `false`

### `--help`
Show help message and exit.

## Special Commands

Inside the interactive REPL, you can use PostgreSQL special commands:

- `\d [pattern]` - List tables, views, and sequences
- `\l` - List databases
- `\dt` - List tables only
- `\q` - Quit the application
- And many more standard PostgreSQL backslash commands

## Environment Variables

You can use standard PostgreSQL environment variables:

- `PGHOST` - Database host
- `PGPORT` - Database port
- `PGUSER` - Database user
- `PGDATABASE` - Database name
- `PGPASSWORD` - Database password (not recommended for security reasons)
