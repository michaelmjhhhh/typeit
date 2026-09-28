# Go port verification

Verified from the project root on macOS ARM64 with Go 1.26.1.

- Go sources, tests, and documentation live directly under the project root.
- Obsolete Rust application/build/release files and unused demo media have been removed.
- All 20 native grammars download from the checksummed `scripts/grammars.json` manifest.
- Generated C sources are excluded from Git; fresh checkout preparation,
  offline cache reuse, and recovery of deleted parser sources are tested.
- All 269 upstream golden snapshots remain under `tests/upstream`.
- `make check`: passed formatting, vet, all Go tests, and the race detector.
- `make build`: passed; executable at `bin/gittype`.
- `bin/gittype --version`: `gittype 0.10.2-go`.
- Terminal smoke: passed at 110×32; three completed stages, zero mistakes,
  100% accuracy, persisted stage/session records, history/details navigation,
  analytics/settings navigation, cursor and alternate-screen restoration.
- Original-fixture comparisons: 112 typing/display snapshots and 157 extraction snapshots.

The terminal walkthrough is recorded in [terminal-smoke.cast](terminal-smoke.cast).
See [migration notes](../GO_PORT.md) for compatibility decisions and verification limits.
