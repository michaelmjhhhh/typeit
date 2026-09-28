# Go migration

The Go application is rooted directly in this directory, with its entry point at
`cmd/typeit`. It uses Bubble Tea for the event loop, Bubbles for loading state,
and Lip Gloss for terminal styling. Rust is not needed to build or run it.

```sh
make build
./bin/typeit /path/to/project
make check
python3 scripts/terminal_smoke.py --record docs/qa/terminal-smoke.cast
```

Go 1.26.1+, Python 3.9+, a C compiler, and Git are required. Tree-sitter and SQLite use CGo.
The supplied executable is a local macOS ARM64 build; build from source for other
platforms with a suitable C toolchain.

## Implemented coverage

- [x] All 20 language grammars, original extraction queries, and capture filters
- [x] File selection, language aliases, size limits, nested and inherited ignore
  files, global Git ignores, and repository excludes
- [x] Five difficulty levels, original truncation rules, and four-line source context
- [x] Unicode typing and display mapping, comments, indentation, mistakes,
  pause/resume, and countdown
- [x] Three-stage sessions, three skips, interrupted sessions, stage results,
  session results, rank animations, and aggregate results
- [x] Original scoring rules, all 63 rank boundaries, and rank artwork
- [x] SQLite history and stage details using the original normalized schema
- [x] Analytics, record sorting and date filters, personal bests, and JSON/CSV export
- [x] Local projects, remote Git clones, repository catalogs, caches, and trending
- [x] Original built-in themes, light/dark mode, custom themes, and save/cancel
- [x] Help, browser share drafts, optional version checks, and terminal restoration

## Verification

`make check` checks Go formatting, runs `go vet ./...`, `go test ./...`, and
`go test -race ./tests/go/...`. The suite includes:

- 112 original Rust typing snapshots, compared for typing and display text.
- 157 original Rust extraction snapshots, compared for extracted code, source
  boundaries, chunk counts, and comment ranges.
- All 20 grammars, all 63 rank boundaries and artwork entries, Unicode input,
  score golden cases, and subsecond scoring behavior.
- Pause timing, skips, session completion and aborts, settings save/cancel,
  coalesced keyboard messages, and terminal screen bounds.
- SQLite compatibility, source and ignore-rule cache invalidation, real Git
  cloning from a local HTTP fixture, and mocked trending/version service caches.

The terminal smoke script runs the actual executable through a pseudo-terminal,
types three complete stages, checks saved SQLite metrics, opens history and
session details, visits analytics and settings, and verifies cursor and alternate
screen restoration. Its recording is in `docs/qa/terminal-smoke.cast` and can be
replayed with an asciicast-compatible player.

Verification was performed on macOS ARM64. Automated comparisons cover the
upstream behavioral fixtures; they do not establish pixel-for-pixel equivalence
between Ratatui and Lip Gloss on every terminal. Linux and Windows execution and
live external-service availability have not been independently verified here.

## Compatibility decisions

- Historical sessions remain in the upstream-compatible schema in `typeit.db`. Config and
  custom-theme JSON retain the original format. Existing repository clones are
  reused under `repos/<host>/<owner>/<repository>`.
- Disposable Rust bincode challenge caches are rebuilt into a separate compressed
  JSON cache under `challenge-cache/`. Source contents, language filters, and
  ignore rules determine reuse.
- Export now produces JSON or CSV; the original command was a placeholder.
- The terminal interface is implemented with Charmbracelet components and retains
  the original artwork, themes, typing rules, and primary navigation workflows.
- Original golden snapshots are retained in `tests/upstream`. The Rust project,
  obsolete release workflows, and one-time conversion scripts have been removed.
  Go checks are run through `make check`.

The upstream MIT license and third-party notices are retained. Generated C
parsers are pinned in `scripts/grammars.json`, carrying forward the original
versions and SHA-256 checksums. `scripts/port_grammars.py` verifies every download. Parser sources are downloaded at build time and ignored by Git. Small bindings
and parser licenses remain in the repository;
Go dependency notices are in `docs/THIRD_PARTY_NOTICES.md`.
