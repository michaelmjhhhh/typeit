# Development Rules

Typeit is a Go rewrite of https://github.com/unhappychoice/gittype using Bubble
Tea, Bubbles, and Lip Gloss. Preserve the upstream MIT notice in LICENSE and
attribution in NOTICE, including the licenses of native grammars and dependencies.

## Layout

- `cmd/typeit`: CLI entry point.
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

Use `TYPEIT_DATA_DIR` to isolate development history. Native grammars are pinned
in `scripts/grammars.json`; Make targets prepare them automatically. Use
`make grammars` before direct Go commands or `make generate` to refresh sources.
Never commit generated native C sources; commit only bindings and licenses.
Do not hand-edit parser C files. Keep meaningful regression tests with behavior changes.

## Test

Do not write any unit tests for TUI changes. Instead, give the user a TODO checklist to manually verify the TUI changes. Give the user the CLI command to open the updated TUI. 

## Conversational Style

- Keep answers short and concise
- No emojis in commits, issues, PR comments, or code
- Technical prose only, be direct
- Use concise, clear, simple language. Define unavoidable jargon before using it.
- Explain non-trivial designs and problems as: problem, concrete example or short trace, then solution. State why the solution is necessary and distinguish it from optional complexity.
- Prefer concrete behavior and small illustrations over abstract summaries, dense terminology, or unexplained lists of changes.
- When the user asks a question, answer it first before making edits or running implementation commands.
- When responding to user feedback or an analysis, explicitly say whether you agree or disagree before saying what you changed.

## Code Quality
 
- Read files in full before applying wide-ranging changes, before editing files you have not fully inspected, and when
asked to investigate or audit. Do not rely on search snippets for broad changes.

## Changelog

Always maintain/update the CHANGELOG for changes. 

## User Override

If the user's instructions conflict with any rule in this document, ask for explicit confirmation before overriding. Only then execute their instructions.