---
title: Features
description: An overview of what pgxcli brings to your PostgreSQL workflow.
keywords: [postgresql syntax highlighting, sql autocompletion, psql features, postgres auto complete, chroma]
sidebar_position: 6
---

pgxcli is an interactive PostgreSQL REPL built in Go. This page highlights the core features that make it fast, readable, and comfortable to use.

---

## Syntax Highlighting

pgxcli applies real-time syntax highlighting to your SQL as you type, using [Chroma](https://github.com/alecthomas/chroma) — the same engine that powers many static site generators and code tools.

![pgxcli syntax highlighting — monokai and onedark themes](/img/syntax.png)

Colors are applied to keywords, identifiers, string literals, operators, and comments, so long queries remain easy to read at a glance.

pgxcli automatically detects your terminal's color depth — TrueColor, 256-color, or 16-color — and picks the right formatter for your environment.

---

## Autocompletion

pgxcli offers word-based completion for SQL keywords. Press `Tab` to cycle through suggestions.

![pgxcli autocompletion in action](/img/completion.gif)
