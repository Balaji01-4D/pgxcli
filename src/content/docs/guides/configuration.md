---
title: Configuration
description: Customize pgxcli with a simple TOML config file.
---

On first run, pgxcli creates a config file at:

```
~/.config/pgxcli/config.toml
```

(Or the OS-equivalent user config directory.)

Every setting has a sensible default. Edit the file to make it yours.

---

## Full Default Config

Here's the complete default configuration:

```toml
[main]
# Postgres prompt
# \t - Current date and time
# \u - Username
# \h - Short hostname (up to first '.')
# \H - Full hostname
# \d - Database name
# \p - Port
# \n - Newline
prompt = "\\u@\\h:\\d> "

# Syntax highlighting style
# Available: monokai, dracula, nord, onedark, github-dark,
#            gruvbox, solarized-dark, xcode-dark, darcula, catppuccin-mocha
# Full list: https://xyproto.github.io/splash/docs/index.html
style = "monokai"

# History file location ("default" = ~/.pgxcli_history.jsonl)
history_file = "default"

# Log file location ("default" = OS-standard location)
log_file = "default"

# Pager behavior for long output
#   auto   - page only when output exceeds terminal height
#   always - always pipe through pager
#   never  - print directly
pager = "auto"

# Error handling for multi-statement execution
#   STOP   - stop on first error
#   RESUME - continue executing remaining statements
on_error = "STOP"
```

---

## Settings Reference

### `prompt`

The string displayed before each input line. Supports these variables:

| Variable | Replaced With |
|----------|---------------|
| `\u` | Current username |
| `\h` | Short hostname (up to first `.`) |
| `\H` | Full hostname |
| `\d` | Current database name |
| `\p` | Port number |
| `\t` | Current date and time |
| `\n` | Newline |

**Default:** `\u@\h:\d> ` → looks like `postgres@localhost:mydb> `

### `style`

The Chroma syntax highlighting theme applied to your SQL as you type.

Some popular choices:

- `monokai` (default)
- `dracula`
- `nord`
- `catppuccin-mocha`
- `github-dark`
- `gruvbox`
- `solarized-dark`

pgxcli automatically detects your terminal's color support (TrueColor, 256-color, or 16-color) and picks the right formatter.

Browse all available styles at [xyproto.github.io/splash/docs](https://xyproto.github.io/splash/docs/index.html).

### `history_file`

Where command history is stored. Set to `"default"` to use `~/.pgxcli_history.jsonl`.

History is saved as JSON Lines. Up to 1000 entries are kept.

### `log_file`

Where debug logs are written. Set to `"default"` for the OS-standard location.

To see debug output, start pgxcli with `--debug`.

### `pager`

Controls when long output is piped through a pager (`less` on Linux/macOS, `more` on Windows).

| Value | Behavior |
|-------|----------|
| `auto` | Page when output exceeds terminal height or 4 KB |
| `always` | Always use the pager |
| `never` | Print directly to the terminal |

**Default:** `auto`

:::tip
Set the `PAGER` environment variable to use a custom pager command, e.g. `PAGER="less -S"`.
:::

### `on_error`

What happens when you run multiple SQL statements and one fails.

| Value | Behavior |
|-------|----------|
| `STOP` | Stop immediately — skip remaining statements |
| `RESUME` | Keep going — execute the rest |

**Default:** `STOP`

For example, if you paste three statements and the second one has a syntax error:

- **STOP**: only the first statement runs.
- **RESUME**: the first and third statements run.
