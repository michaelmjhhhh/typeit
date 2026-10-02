# Changelog

## Unreleased

## 0.0.1 - 2026-10-03

### Changed

- Replace the previous built-in themes with Charm, Dracula, Catppuccin, Base 16,
  and Default. Removed preset IDs fall back to Charm; custom palettes are preserved.
- Update the title artwork with Typeit branding.
- Open history details from the already loaded session data instead of querying
  the same stages again. Refresh history with R to load updated records.

### Removed

- Remove the unused game-mode selection module and its dedicated test.
- Remove the forwarding repository parser and unused source-scan progress callback.

### Fixed

- Restore theme colors after nested text styles reset them, preventing the
  terminal's background from showing through title and typing screens while
  preserving cursor and mistake highlights.
