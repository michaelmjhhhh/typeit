# Typeit

Turn source code from real repositories into terminal typing challenges.

This is a Go port of [unhappychoice/gittype](https://github.com/unhappychoice/gittype),
using **Bubble Tea**, **Bubbles**, and **Lip Gloss**. It retains the upstream MIT
license, Tree-sitter language grammars, themes, rank artwork, and scoring rules.

## Build and run

Requirements: Go 1.26.1 or newer, Python 3.9+, a C compiler, and Git. Tree-sitter and SQLite use
CGo; Rust is not required to build or run the Go executable.

```sh
make build
./bin/gittype /path/to/your/project
```

The first `make` command downloads the pinned native parser sources and verifies
their SHA-256 checksums. Later runs reuse verified local files. Generated C
sources are ignored by Git; they are not part of the repository.

To run directly after preparing the grammars:

```sh
make grammars
go run ./cmd/gittype . --langs go,rust
```

Install the executable into your Go binary directory:

```sh
make grammars
go install ./cmd/gittype
```

## Features

- Tree-sitter extraction for Rust, TypeScript/TSX, JavaScript/JSX, Python, Go,
  Ruby, Swift, Kotlin, Java, PHP, C#, C, C++, Haskell, Dart, Scala, Clojure,
  Elixir, Erlang, and Zig.
- Easy, Normal, Hard, Wild, and Zen challenges, including complete files.
- Unicode input, automatic indentation and comment skipping, live metrics,
  a countdown, pause/resume, three-stage sessions, and three skips per session.
- The original scoring formula, 63 ranks, ASCII artwork, rank animations,
  personal bests, stage results, session results, and total results.
- SQLite history, sortable records, date filters, session details, and four
  analytics views.
- Local projects, remote repositories, repository caches, and trending discovery.
- Fifteen built-in themes, a custom theme, and light/dark previews with save/cancel.
- JSON/CSV exports and browser share drafts.

## Commands

```sh
gittype                                  # Current directory
gittype /path/to/project --langs go,rust
gittype --repo owner/repository
gittype --repo https://github.com/owner/repository
gittype --repo git@github.com:owner/repository.git

gittype history
gittype stats
gittype export --format json --output sessions.json
gittype export --format csv --output sessions.csv

gittype cache stats
gittype cache list
gittype cache clear
gittype repo list
gittype repo play
gittype repo clear                        # Asks before clearing
gittype repo clear --force

gittype trending
gittype trending rust --period weekly
gittype trending rust owner/repository
```

The game and interactive views require a terminal. Export, cache commands, and
repository listing also work in scripts. `--help` lists the available options.

## Controls

On the title screen, use **←/→** or **H/L** to select a difficulty, then **Space**
to select a challenge. Press **Space** again to start the countdown. Type the
code as displayed; comments, blank lines, and leading indentation are automatic.
Mistakes leave the cursor in place. Backspace does not undo progress.

**Esc** pauses. While paused, **S** skips and **Q** ends the session. Other keys
resume without entering a character. **Ctrl+C** exits. Complete three stages
and press **Space** on the last stage result to see the rank animation and
session summary; **S** skips the animation.

From the title: **R** opens records, **A** analytics, **S** settings, and **I/?**
help. In settings, **←/→** chooses a section, **↑/↓** previews a choice, **Space**
saves, and **Esc** cancels.

## Data and compatibility

Data defaults to `~/.gittype`. Set `GITTYPE_DATA_DIR` to isolate a run:

```sh
GITTYPE_DATA_DIR=/tmp/gittype-demo ./bin/gittype .
```

The Go port reads and writes the original `gittype.db` schema and `config.json`
format, and reuses clones under `repos/<host>/<owner>/<repository>`. Custom
colors live in `custom-theme.json`; additional named themes can be placed in
`themes/*.json`.

Extracted challenges use a separate, compressed JSON cache under
`challenge-cache/`. It checks source contents and language filters before reuse.
Rust's disposable bincode challenge caches are rebuilt on the first Go run.
No historical sessions need conversion.

Use `.gitignore`, `.ignore`, and `.gittypeignore` to exclude files. Global Git
ignores and repository excludes are also respected. Generated/build directories,
symlinks, binary files, and files larger than 1 MiB are excluded from extraction.

## Verification and development

```sh
make check
make build
python3 scripts/terminal_smoke.py
```

This runs formatting checks, `go vet`, all Go tests, and the race detector. Tests
reuse **112 upstream typing snapshots** and **157 upstream extraction snapshots**,
and cover all grammars, Unicode, scoring boundaries, sessions, persistence,
cache invalidation, Git cloning, service caching, and terminal screen bounds.

The original behavioral snapshots are retained under `tests/upstream`. The Rust
application and obsolete release tooling have been removed. The Go application
does not execute or link Rust code. See
[the migration notes](docs/GO_PORT.md), [architecture](docs/GO_ARCHITECTURE.md),
and [Go contribution guide](docs/GO_CONTRIBUTING.md).

## License

This project is a Go rewrite of
[GitType](https://github.com/unhappychoice/gittype), licensed under [MIT](LICENSE).

- Original GitType material: Copyright (c) 2025 unhappychoice.
- Go rewrite and modifications: Copyright (c) 2026 Typeit contributors.

The upstream copyright and MIT permission notice are retained. The rewrite's
copyright notice applies to its contributions and does not replace the original
author's copyright. See [NOTICE](NOTICE) for attribution and provenance.

Third-party parser licenses are retained under `internal/grammars`; Go dependency
notices are in [THIRD_PARTY_GO.md](THIRD_PARTY_GO.md). Additional upstream dependency and native grammar
notices remain in [LICENSE-THIRD-PARTY](LICENSE-THIRD-PARTY).
