<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
![CLI][cli-shield]
![pgxcli][pgxcli-shield]

![Go][go-shield]
![PostgreSQL][postgres-shield]
<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://pgxcli.vercel.app/">
    <img src="https://res.cloudinary.com/dsdupsv2g/image/upload/q_auto/f_auto/v1777298704/1_oepvrn.png" alt="pgxcli banner" width="100%"/>
  </a>
  <h1 align="center">pgxcli</h1>
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

`pgxcli` is an interactive PostgreSQL command-line client built in Go, designed for a fast, and smooth REPL experience. It includes syntax highlighting, keyword autocompletion, command history, and support for PostgreSQL backslash commands.

Highlights:
* Interactive REPL with customizable prompt and style.
* SQL syntax highlighting.
* SQL keyword autocompletion.
* Persistent command history.
* PostgreSQL special backslash commands (`\d`, `\l`, `\dt`, `\q`).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- PGXCLI VS PGCLI -->
## Comparison with pgcli

[Pgcli][pgcli-url] is a mature PostgreSQL CLI developed over many years, which has set the standard for interactive PostgreSQL clients.

**pgxcli** takes the simpler approach, focusing on speed and minmal setup. It is a singe Go binary with fast startup and TOML configuration. If you need a lightweight, It may be good fit. for a more feature-rich, established experience, pgcli remains the benchmark.

### Where pgxcli stands out:
#### Now 
* Single binary, no external runtime dependencies
* Fast startup and better performance

#### Planned
* Modern CLI Interface
* Streaming query results for large tables
* Browser based Table view via localhost
* Performance improvements for large tables
* Direct Table export to SQL INSERT statements, CSV, MD tables, Excel, and HTML.

<details>
  <summary><strong>Which one should I use?</strong></summary>

Right now, I would definitely choose pgcli. I think no explanation is needed.

That could change as pgxcli matures. I would really appreciate if you give pgxcli a try and share your feedback. If you want to contribute, that would be even better.

</details>

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

### Installation

#### Linux
```bash
# Debian / Ubuntu
sudo dpkg -i pgxcli_*_linux_amd64.deb
```

#### macOS
```bash
brew tap Balaji01-4D/pgxcli
brew install pgxcli
```

#### Windows
Download the `.msi` from the [releases page][releases-url] and run it.

For more installation methods, see the [Installation Guide][install-ref].

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

<img src="https://res.cloudinary.com/dsdupsv2g/image/upload/q_auto/f_auto/v1777298704/5_h2fxui.png" alt="pgxcli flags screenshot" width="100%"/>

For full flag documentation, see the [docs][cli-ref].

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONFIGURATION -->
## Configuration

On first run, a config file is created at:

* `~/.config/pgxcli/config.toml` (or the OS-equivalent user config directory)

For configuration documentation, see the [docs][config-ref]

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [pgx][pgx-url]
* [Cobra][cobra-url]
* [Viper][viper-url]
* [go-pretty][go-pretty-url]
* [go-prompter][go-prompter-url]
* Inspired by [pgcli][pgcli-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[cli-shield]: https://img.shields.io/badge/CLI-%23000000?style=for-the-badge&logo=iterm2&logoColor=white
[pgxcli-shield]: https://img.shields.io/badge/Pgxcli-7B36ED?style=for-the-badge&logo=database&logoColor=white
[go-shield]: https://img.shields.io/badge/Go-%23000000?style=for-the-badge&logo=go&logoColor=white
[postgres-shield]: https://img.shields.io/badge/PostgreSQL-7B36ED.svg?style=for-the-badge&logoColor=white

[pgx-url]: https://github.com/jackc/pgx
[cobra-url]: https://github.com/spf13/cobra
[viper-url]: https://github.com/spf13/viper
[go-pretty-url]: https://github.com/jedib0t/go-pretty
[go-prompter-url]: https://github.com/jedib0t/go-prompter
[pgcli-url]: https://github.com/dbcli/pgcli
[cli-ref]: https://pgxcli.vercel.app/reference/cli-reference/
[config-ref]: https://pgxcli.vercel.app/guides/configuration/
[install-ref]: https://pgxcli.vercel.app/docs/guides/getting-started#installation
[releases-url]: https://github.com/Balaji01-4D/pgxcli/releases
