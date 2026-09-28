# Installation

Prebuilt binaries include the native parsers and SQLite. Go, Python, and a C
compiler are only needed when building from source. Git is needed for cloning
remote repositories. Supported releases: macOS 14+ (Intel / Apple Silicon),
Linux (x86-64 / ARM64), and Windows x86-64.

## macOS and Linux

```sh
curl -fsSL https://github.com/michaelmjhhhh/typeit/releases/latest/download/install.sh | sh
```

The installer verifies the archive's SHA-256 checksum and installs `typeit` into
`~/.local/bin`, without sudo. If needed, add this directory to your shell profile:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

To inspect the installer before running it, download `install.sh` from Releases.
Run it with `sh install.sh`. For a specific version or location:

```sh
TYPEIT_INSTALL_DIR="$HOME/bin" sh install.sh --version v0.1.0
```

Running the installer again updates the executable. License notices are stored
under `${XDG_DATA_HOME:-$HOME/.local/share}/typeit`.

## Windows and manual installation

Download the matching archive and `SHA256SUMS` from
[Releases](https://github.com/michaelmjhhhh/typeit/releases/latest). Extract it,
then run `typeit` or `typeit.exe` in a terminal. Optionally put the executable in
a directory on your PATH. Keep the included license files with redistributed
copies.

Windows checksums can be checked with `Get-FileHash <archive> -Algorithm SHA256`.
On Linux use `sha256sum`; on macOS use `shasum -a 256`.

## Data and removal

Data defaults to `~/.typeit`: `typeit.db` holds history and `config.json` holds
settings. Use `TYPEIT_DATA_DIR` to select another directory.

On first use, if `~/.typeit` does not exist, Typeit copies history and theme
settings from `~/.gittype`, leaving the originals intact. Existing Typeit data
is never merged or overwritten. Downloaded repositories and caches are rebuilt
as needed. In a custom directory, an existing `gittype.db` is copied to
`typeit.db` once. `GITTYPE_DATA_DIR` is still accepted as a fallback when
`TYPEIT_DATA_DIR` is unset.

Uninstall by removing the executable. This leaves your typing history intact.

## Building releases

The release workflow uses native runners for five OS/architecture combinations.
Each runs tests and builds a versioned binary, with terminal smoke tests on
macOS/Linux. Linux release builds use musl for static linking. Windows uses
MinGW; macOS uses Apple's compiler.

A manual workflow run builds downloadable Actions artifacts. Pushing a `vX.Y.Z`
tag publishes the archives, installer, checksums, and license notices after every
platform succeeds. Use `scripts/package_release.py --version vX.Y.Z` for a local
native build. Use `scripts/release_checksums.py` after collecting all five archives.
