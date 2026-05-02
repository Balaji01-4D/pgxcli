---
title: Environment Variables
description: Environment variables that pgxcli reads for connection defaults and configuration.
keywords: [postgres environment variables, PGPASSWORD, PGUSER, libpq env, PGHOST, PGDATABASE]
sidebar_position: 2
---

pgxcli reads several environment variables to fill in connection defaults and configure behavior. You never _have_ to set any of them — but they're useful for scripting or when you connect to the same server repeatedly.

---

## pgxcli-Specific

These take priority over the standard PostgreSQL variables:

| Variable | Description |
|----------|-------------|
| `PGXUSER` | Default username for connections |
| `PGXPASSWORD` | Password used when `--no-password` (`-w`) is set |

---

## Standard PostgreSQL

pgxcli also respects the standard `libpq` environment variables:

| Variable | Description |
|----------|-------------|
| `PGUSER` | Default username (used if `PGXUSER` is not set) |
| `PGPASSWORD` | Default password (used if `PGXPASSWORD` is not set) |
| `PGHOST` | Default host address |
| `PGPORT` | Default port number |
| `PGDATABASE` | Default database name |

These work exactly like they do with `psql` or any other `libpq`-based tool.

---

## Pager

| Variable | Description |
|----------|-------------|
| `PAGER` | Custom pager command (default: `less` on Linux/macOS, `more` on Windows) |

If `PAGER` is not set, pgxcli uses `less` with `-SRFX` flags on Unix systems.

:::tip
Set `PAGER="less -S"` to disable line wrapping in paged output — useful for wide tables.
:::

---

## Resolution Priority

When multiple sources provide the same value, pgxcli resolves them in this order (highest priority first):

### Username
1. `--username` / `-u` / `--user` / `-U` flag
2. Positional argument
3. `PGXUSER` environment variable
4. `PGUSER` environment variable
5. Current OS username

### Password
1. Interactive form input (with `-i`)
2. `--password` / `-W` prompt
3. `PGXPASSWORD` environment variable
4. `PGPASSWORD` environment variable
5. Auto-prompt on authentication failure (unless `-w` is set)
