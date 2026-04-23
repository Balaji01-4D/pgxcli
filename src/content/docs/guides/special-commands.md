---
title: Special Commands
description: Backslash commands and built-in commands available inside the pgxcli REPL.
---

pgxcli supports PostgreSQL-style backslash commands and a few built-in extras. Type them directly at the prompt.

---

## Session Commands

These control your session:

| Command | Description |
|---------|-------------|
| `\q` | Quit pgxcli |
| `\c <database>` | Switch to a different database on the same server |
| `\connect <database>` | Same as `\c` |
| `\conninfo` | Show current connection details (database, user, host, port) |

### Switching Databases

```
\c other_db
```

pgxcli closes the current connection and opens a new one to `other_db`. Your prompt updates to reflect the new database. The server, user, and port stay the same.

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
| `clear` | Clear the terminal screen |

---

## SQL Execution

Anything that isn't a special command is treated as SQL and sent to the database.

You can run multiple statements at once — pgxcli splits them by semicolons and executes each one in order. If a statement fails, the behavior depends on your [`on_error`](/pgxcli/guides/configuration/#on_error) setting.

Query results are rendered as tables. Write operations (like `INSERT` or `CREATE TABLE`) print the command tag (e.g., `INSERT 0 1`). Execution time is shown after each result.
