# Contributor guidelines

Typeit is a Go rewrite of https://github.com/unhappychoice/gittype using Bubble
Tea, Bubbles, and Lip Gloss. Preserve the upstream MIT notice in LICENSE and
attribution in NOTICE, including the licenses of native grammars and dependencies.

## Layout

- `cmd/gittype`: CLI entry point.
- `internal/domain`: pure typing, challenge, scoring, and analytics rules.
- `internal/infra`: parsing, persistence, files, Git, themes, and external services.
- `internal/ui`: Bubble Tea state, commands, input handling, and views.
- `internal/cli`: argument parsing and command dispatch.
- `internal/grammars`: CGo bindings, parser licenses, and ignored generated C sources.
- `tests/go`: external Go tests; keep tests outside production packages.
- `tests/upstream`: original MIT-licensed golden snapshots; do not blindly update.

Keep the domain independent of UI and infrastructure. Perform slow I/O through
Bubble Tea commands, capture immutable inputs, and update state through messages.
See `docs/GO_ARCHITECTURE.md` and `docs/GO_CONTRIBUTING.md` for further guidance.

## Verification

Use Go 1.26.1+, Python 3.9+, a C compiler, and Git. Before committing or pushing, run:

```sh
make check
make build
python3 scripts/terminal_smoke.py
```

Use `GITTYPE_DATA_DIR` to isolate development history. Native grammars are pinned
in `scripts/grammars.json`; Make targets prepare them automatically. Use
`make grammars` before direct Go commands or `make generate` to refresh sources.
Never commit generated native C sources; commit only bindings and licenses.
Do not hand-edit parser C files. Keep meaningful regression tests with behavior changes.

## Repository conventions

Write code, documentation, comments, and commit messages in English. Prefer small,
focused functions, explicit error handling, and Conventional Commit messages.
Preserve others' changes. Never bypass hooks or force-push `main`. Do not publish
or modify shared remote state without user authorization. Confirm changes to CI
workflows unless already covered by the user's request. For UI changes, include
a screenshot or terminal recording in a pull request.
