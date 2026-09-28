# Contributing to the Go port

Use Go 1.26.1+, Python 3.9+, a C compiler, and Git. Build with `make build` and run the executable
from `bin/gittype`. Use `GITTYPE_DATA_DIR` for development and tests to avoid
mixing practice sessions with your normal history.

Run `make check` before submitting changes. Keep Go tests under `tests/go`, using
external test packages. The original Rust snapshots are read-only golden inputs;
do not update them simply to make a Go test pass. Add focused tests for changed
behavior and preserve the domain's independence from UI and infrastructure.

Native parsers are downloaded automatically by the Make targets. Use
`make grammars` before direct `go` commands, or `make generate` to refresh them.
The grammar script downloads the versions recorded in `scripts/grammars.json`,
verifies SHA-256 checksums, retains license files, and writes CGo bindings.
Generated C sources are ignored by Git. Commit only their version/checksum
manifest, small CGo bindings, and licenses. Do not hand-edit generated parser C
files. When changing a grammar version, verify its binding and parser symbol.

Use Bubble Tea commands for slow operations. Capture immutable inputs before
starting a command, and mutate model state only while handling messages. Check
errors from persistence and network operations. Keep rendering independent of
network access, clocks, and filesystem reads outside embedded assets.

`GO_ARCHITECTURE.md` describes the Go implementation. Follow `AGENTS.md` for
repository conventions, verification, license preservation, and Git workflow.
Assets, extraction queries, language metadata, and the SQL schema are maintained
as checked-in Go or JSON/SQL sources; they do not require a Rust checkout.
