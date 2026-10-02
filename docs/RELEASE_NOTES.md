Typeit turns source code into terminal typing challenges.

## v0.0.1

- Replace the previous built-in themes with Charm, Dracula, Catppuccin, Base 16,
  and Default, and update the Typeit logo. Removed preset IDs fall back to Charm;
  custom palettes remain intact.
- Preserve theme backgrounds through nested text styling while keeping cursor
  and mistake highlights visible.
- Open history details from already loaded session data. Press R in history to
  refresh records.
- Remove unused internal game-mode selection, the forwarding repository parser,
  and the unused source-scan progress callback.

Download a native executable for macOS (Intel / Apple Silicon), Linux (x86-64 / ARM64), or Windows x86-64. Go, Python, and a compiler are not required to run it. Git is needed only for remote repository features.

### macOS / Linux

```sh
curl -fsSL https://github.com/michaelmjhhhh/typeit/releases/latest/download/install.sh | sh
```

### Windows

Extract `typeit_windows_amd64.zip` and run `typeit.exe` in a terminal.

The archives include the MIT license, upstream attribution, and third-party notices. `SHA256SUMS` provides archive checksums. Linux builds are statically linked; macOS builds target macOS 14 or later.
