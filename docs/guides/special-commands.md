---
title: Special Commands
description: Backslash commands and built-in commands available inside the pgxcli REPL.
sidebar_position: 5
---

pgxcli supports PostgreSQL-style backslash commands. Type them directly at the prompt.

---

## Session Commands

These control your session:

| Command | Description |
|---------|-------------|
| `\q`, `\quit`, `\exit` | Quit pgxcli (case-insensitive) |
| `\c <database>` | Switch to a different database on the same server |
| `\connect <database>` | Same as `\c` |
| `\conninfo` | Show current connection details (database, user, host, port) |

### Switching Databases

```
\c other_db
```

pgxcli closes the current connection and opens a new one to `other_db`. The server, user, and port stay the same.

### Connection Info

```
\conninfo
```

Outputs something like:

```
You are connected to database "mydb" as user "postgres" on Host "localhost" at port 5432
```

---

## Catalog Commands

These come from the [pgxspecial](https://github.com/balaji01-4d/pgxspecial) library and work like their `psql` equivalents:

| Command | Description |
|---------|-------------|
| `\d [pattern]` | Describe a table, view, or other object |
| `\dt [pattern]` | List tables |
| `\dv [pattern]` | List views |
| `\di [pattern]` | List indexes |
| `\ds [pattern]` | List sequences |
| `\df [pattern]` | List functions |
| `\l` | List all databases |
| `\dn` | List schemas |
| `\du` | List roles |
| `\dx` | List installed extensions |

:::tip
Commands with `[pattern]` accept an optional filter. For example, `\dt public.*` lists only tables in the `public` schema.
:::

---

## Built-in Commands

These are pgxcli-specific:

| Command | Description |
|---------|-------------|
| `\clear` | Clear the terminal screen |

---

## SQL Execution

Anything that isn't a special command is treated as SQL and sent to the database.
