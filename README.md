# Typeit

Practice typing with real code from your projects or GitHub repositories.
A Go rewrite of [GitType](https://github.com/unhappychoice/gittype), built with
[Charmbracelet](https://charm.sh/).

## Install

**macOS / Linux** — no Go, Python, or compiler needed:

```sh
curl -fsSL https://github.com/michaelmjhhhh/typeit/releases/latest/download/install.sh | sh
```

The installer chooses your architecture, verifies SHA-256, and installs to
`~/.local/bin`. Follow its PATH instruction if that directory is not on your PATH.
Run the same command to update.

**Windows** — download `typeit_windows_amd64.zip` from
[Releases](https://github.com/michaelmjhhhh/typeit/releases/latest), extract it,
and run `typeit.exe` in a terminal. macOS/Linux archives are also available there.

## Use

```sh
typeit                              # Practice with the current directory
typeit /path/to/project --langs go
typeit --repo owner/repository      # Requires Git
typeit history                      # Past sessions
typeit --help                       # All commands
```

Choose a difficulty with **←/→**, press **Space** to begin, and **Esc** to pause.
Includes 20 languages, five difficulties, [Charmbracelet-inspired themes](docs/THEMES.md), scores, and session history.

## Develop

Requires Go 1.26.1+, Python 3.9+, a C compiler, and Git.

```sh
make build
./bin/typeit .
make check
```

Native parsers are downloaded and verified automatically; generated sources stay
out of Git. See [contributing](docs/GO_CONTRIBUTING.md) and
[installation options](docs/INSTALLATION.md).

## License

[MIT](LICENSE). Original GitType © 2025 unhappychoice; Go rewrite © 2026 Typeit
contributors. [Attribution](NOTICE) · [Third-party notices](docs/THIRD_PARTY_NOTICES.md).
