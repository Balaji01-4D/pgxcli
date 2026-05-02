---
title: Configuration
description: Customize pgxcli with a simple TOML config file.
keywords: [pgxcli config, toml configuration, postgres theme, psql pager, syntax highlighting theme]
sidebar_position: 4
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
# \h - Short hostname of the server (up to first '.')
# \H - Hostname of the server
# \d - Database name
# \p - Database port
# \n - Newline
prompt = "\\u@\\h:\\d> "

# Syntax highlighting style
# Available:
#   default (same as monokai), monokai, dracula, nord, onedark, github-dark
#   github, gruvbox, gruvbox-light, solarized-dark, solarized-light
#   catppuccin-mocha, rose-pine, tokyonight-night, xcode-dark, ...
# Full list: https://xyproto.github.io/splash/docs/index.html
style = "monokai"

# History file location ("default" = ~/.pgxcli_history.jsonl)
history_file = "default"

# Log file location ("default" = OS-standard location)
log_file = "default"

# Pager behavior for long output
#   auto   - use pager only when output is large
#   always - always use pager in interactive terminal mode
#   never  - never use pager
pager = "auto"

# Error handling for multi-statement execution
#   STOP   - stop on first error
#   RESUME - continue executing remaining statements
on_error = "STOP"

# Table border/line style
# Valid values:
#   "none", "ascii", "light", "heavy", "double", "double_long"
#   "light_heavy", "heavy_light", "light_double", "double_light", "rounded", "markdown"
#   "graphical", "merger", "default", "dotted", "arrow", "starry", "hearts"
#   "circuit", "nature", "artistic", "8bit", "chaos", "dots", "blocks", "zen"
#   "vintage", "sketch", "arrow_double", "celestial", "cyber", "runic", "industrial"
#   "ink", "arcade", "blossom", "frosted", "mosaic", "ufo", "steampunk", "galaxy"
#   "jazz", "puzzle", "hypno"
[table]
style = "default"

# Table text colors
# Valid values:
#   "black", "red", "green", "yellow", "blue", "magenta", "cyan", "white", "default"
# High-intensity variants use a trailing "+":
#   "black+", "red+", "green+", "yellow+", "blue+", "magenta+", "cyan+", "white+"
[table.color]
header  = "cyan"
column  = "white"
caption = "white"
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
- `default` (alias for `monokai`)
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
| `auto` | Page when output exceeds terminal height |
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

---

## Table Settings

These settings live under the `[table]` section and control how query results are rendered.

### `table.style`

Controls the border/line style of result tables.

**Default:** `default`

| Group | Styles |
|-------|--------|
| Plain | `none`, `ascii` |
| Single-line | `light`, `heavy`, `rounded`, `dotted`, `default` |
| Double-line | `double`, `double_long`, `double_light` |
| Mixed | `light_heavy`, `heavy_light`, `light_double` |
| Special | `markdown`, `graphical`, `merger` |
| Decorative | `arrow`, `arrow_double`, `starry`, `hearts`, `circuit`, `nature`, `artistic`, `8bit`, `chaos`, `dots`, `blocks`, `zen` |
| Themed | `vintage`, `sketch`, `celestial`, `cyber`, `runic`, `industrial`, `ink`, `arcade`, `blossom`, `frosted`, `mosaic`, `ufo`, `steampunk`, `galaxy`, `jazz`, `puzzle`, `hypno` |

:::tip
Use `markdown` to copy-paste query results directly into Markdown documents.
:::

### `table.color.header`

Foreground color for the header row text.

**Default:** `cyan`

### `table.color.column`

Foreground color for data cell text.

**Default:** `white`

### `table.color.caption`

Foreground color for the caption line (e.g., row count footer).

**Default:** `white`

#### Available Colors

| Value | Description |
|-------|-------------|
| `default` | Terminal default for that element |
| `black` | Black |
| `red` | Red |
| `green` | Green |
| `yellow` | Yellow |
| `blue` | Blue |
| `magenta` | Magenta |
| `cyan` | Cyan |
| `white` | White |
| `black+` | High-intensity black (bright) |
| `red+` | High-intensity red |
| `green+` | High-intensity green |
| `yellow+` | High-intensity yellow |
| `blue+` | High-intensity blue |
| `magenta+` | High-intensity magenta |
| `cyan+` | High-intensity cyan |
| `white+` | High-intensity white (brightest) |
