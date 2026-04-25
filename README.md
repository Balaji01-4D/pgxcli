<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://balaji01-4d.github.io/pgxcli/">
    <img src="https://res.cloudinary.com/dsdupsv2g/image/upload/v1776949930/logo_l1mlz5.png" alt="pgxcli banner" width="420"/>
  </a>
  <h3 align="center">pgxcli</h3>
  <p align="center">
    Interactive PostgreSQL command-line client written in Go.
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li><a href="#comparison-with-pgcli">Comparison with pgcli</a></li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
        <li><a href="#development">Development</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#configuration">Configuration</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project

`pgxcli` is an interactive PostgreSQL command-line client built in Go. It focuses on a fast, friendly REPL experience with syntax highlighting, keyword autocompletion, history, and support for PostgreSQL backslash commands.

Key highlights:
* Interactive REPL with customizable prompt and style.
* SQL syntax highlighting while typing.
* SQL keyword autocompletion.
* Persistent command history.
* PostgreSQL special backslash commands (for example: `\d`, `\l`, `\dt`, `\q`).
* Configurable error behavior for multi-statement execution (`STOP` / `RESUME`).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- PGXCLI VS PGCLI -->
## Comparison with pgcli

[pgcli][pgcli-url] is an excellent, mature PostgreSQL CLI with over a decade of development. It has set the standard for interactive PostgreSQL clients.

**pgxcli brings simplicity and speed to PostgreSQL.** If you want a single Go binary with fast startup and TOML config, it's worth exploring. For mature, production-hardened features, pgcli remains the standard.

**Quick Wins with pgxcli:**
* **Zero dependencies:** A single Go binary, no Python runtime or virtual environments needed.
* **Instant startup:** Near-instantaneous launch times, getting you to the query prompt faster.
* **Simple configuration:** Modern, straightforward TOML setup.

What pgxcli has today:
* Core interactive REPL experience
* PostgreSQL meta-commands
* Syntax highlighting
* Basic autocompletion (keywords)
* History persistence
* Pager support
* TOML configuration

<details>
  <summary><strong>Which one should I use?</strong></summary>

Choose **pgxcli** if you:
* Want a fast, single Go binary without a Python runtime
* Prefer standard TOML configuration
* Value minimal, focused tooling and fast startup times
* Want to contribute to an early-stage Go project

Choose **pgcli** if you:
* Need mature, battle-tested tooling for production environments
* Require advanced features like SSH tunnels and keyring integration
* Need rich, schema-aware completion right now
* Prefer Python-based tools and ecosystem
</details>

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

```sh
# positional arguments
pgxcli mydb myuser

# flags
pgxcli --host localhost --port 5432 --user postgres --dbname postgres

# connection URI
pgxcli postgres://user:password@localhost:5432/dbname

# interactive connection form
pgxcli -i
```

For full flag documentation, see the [docs][https://balaji01-4d.github.io/pgxcli/reference/cli-reference/].

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONFIGURATION -->
## Configuration

On first run, a config file is created at:

* `~/.config/pgxcli/config.toml` (or the OS-equivalent user config directory)

For configuration documentation, see the [docs][https://balaji01-4d.github.io/pgxcli/reference/cli-reference/]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap

See the [open issues][issues-url] for proposed features and known issues.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the MIT License. See [`LICENSE`][license-url] for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [pgx][pgx-url]
* [Cobra][cobra-url]
* [Viper][viper-url]
* [go-pretty][go-pretty-url]
* [go-prompter][go-prompter-url]
* [pg_query_go][pg-query-url]
* Inspired by [pgcli][pgcli-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/Balaji01-4D/pgxcli.svg?style=for-the-badge
[contributors-url]: https://github.com/Balaji01-4D/pgxcli/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/Balaji01-4D/pgxcli.svg?style=for-the-badge
[forks-url]: https://github.com/Balaji01-4D/pgxcli/network/members
[stars-shield]: https://img.shields.io/github/stars/Balaji01-4D/pgxcli.svg?style=for-the-badge
[stars-url]: https://github.com/Balaji01-4D/pgxcli/stargazers
[issues-shield]: https://img.shields.io/github/issues/Balaji01-4D/pgxcli.svg?style=for-the-badge
[issues-url]: https://github.com/Balaji01-4D/pgxcli/issues
[license-shield]: https://img.shields.io/github/license/Balaji01-4D/pgxcli.svg?style=for-the-badge
[license-url]: https://github.com/Balaji01-4D/pgxcli/blob/main/LICENSE

[go-url]: https://go.dev/
[pgx-url]: https://github.com/jackc/pgx
[cobra-url]: https://github.com/spf13/cobra
[viper-url]: https://github.com/spf13/viper
[go-pretty-url]: https://github.com/jedib0t/go-pretty
[go-prompter-url]: https://github.com/jedib0t/go-prompter
[pg-query-url]: https://github.com/pganalyze/pg_query_go
[cli-reference-url]: https://github.com/Balaji01-4D/pgxcli/blob/main/docs/src/content/docs/reference/cli-reference.md
[pgcli-url]: https://github.com/dbcli/pgcli
