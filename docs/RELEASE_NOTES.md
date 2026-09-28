Typeit turns source code into terminal typing challenges.

Typeit now stores history in `~/.typeit/typeit.db` and settings in
`~/.typeit/config.json`. Set `TYPEIT_DATA_DIR` to use a custom location.
On first launch, history and theme settings are copied from `~/.gittype` if
the new directory does not exist. Original files remain intact, and existing
Typeit data is never overwritten. Help text and result sharing use Typeit branding.

Download a native executable for macOS (Intel / Apple Silicon), Linux (x86-64 / ARM64), or Windows x86-64. Go, Python, and a compiler are not required to run it. Git is needed only for remote repository features.

### macOS / Linux

```sh
curl -fsSL https://github.com/michaelmjhhhh/typeit/releases/latest/download/install.sh | sh
```

### Windows

Extract `typeit_windows_amd64.zip` and run `typeit.exe` in a terminal.

The archives include the MIT license, upstream attribution, and third-party notices. `SHA256SUMS` provides archive checksums. Linux builds are statically linked; macOS builds target macOS 14 or later.
