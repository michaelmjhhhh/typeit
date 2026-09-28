# Go architecture

`cmd/typeit` invokes the CLI adapter. The application is written in Go; the
native Tree-sitter grammars and SQLite engine are C dependencies accessed through
CGo, as Tree-sitter grammars were also native dependencies of the Rust version.

```
cmd/typeit/         executable entry point
internal/cli/       argument parsing, command dispatch, export
internal/domain/    typing, challenges, scoring, ranks, totals, analytics
internal/infra/     parsing, SQLite, files, Git, caching, themes, HTTP
internal/ui/        Bubble Tea model, transitions, Lip Gloss views
internal/grammars/  CGo bindings, licenses, and ignored native build inputs
assets/             embedded themes, logos, ranks, digits, animation messages
tests/go/           unit, integration, original-fixture parity tests
scripts/            reproducible asset generation and terminal smoke checks
```

The domain layer has no UI, storage, or network imports. Infrastructure returns
domain values. The UI owns the Bubble Tea model and routes input to the typing
and scoring rules. I/O work runs in `tea.Cmd` functions and returns messages;
views render state without performing I/O. A cancellable context follows remote
requests, cloning, and source scanning.

Typing uses rune indices throughout. A typing-to-source mapping removes comments
and leading/trailing whitespace, while a separate display mapping preserves
comments, indentation, tab arrows, and newline markers. Timings exclude countdown
and pause durations. Stage scores use correct keystrokes for CPM; session scores
preserve the upstream rule that uses all valid-stage keystrokes for CPM. Total
scores sum session scores and include partial effort in aggregate typing metrics.

SQLite writes each session, its challenges, and its stages in one transaction.
Foreign keys and a busy timeout are enabled. SQL timestamps and rank-position
columns match the original schema. Configuration and cache writes use temporary
files followed by an atomic rename.

Each language uses the exact grammar version and checksum from `scripts/grammars.json`.
The original extraction queries, capture-kind mappings, difficulty limits,
and rank assets are retained in the Go sources and embedded data. Native
parser C sources are downloaded by `make grammars` and excluded from version
control. The Make build/test targets prepare them automatically; direct `go`
commands require this preparation once. Python 3.9+ and network access are needed
for the initial download. Subsequent runs verify and reuse the local sources.

Original golden snapshots are retained in `tests/upstream` for comparison.
The Rust application, build files, and release tooling are not required.
