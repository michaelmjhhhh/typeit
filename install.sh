#!/bin/sh
set -eu

fail() { printf 'typeit: %s\n' "$*" >&2; exit 1; }
version=latest
while [ "$#" -gt 0 ]; do
    case "$1" in
        --version) [ "$#" -ge 2 ] || fail '--version requires a value'; version=$2; shift 2 ;;
        --help|-h) printf 'Usage: install.sh [--version vX.Y.Z]\nSet TYPEIT_INSTALL_DIR to change ~/.local/bin.\n'; exit 0 ;;
        *) fail "Unknown argument: $1" ;;
    esac
done
case "$version" in
    latest) ;;
    *) printf '%s\n' "$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$' || fail 'Version must be vX.Y.Z' ;;
esac
case "$(uname -s)" in Darwin) os=darwin ;; Linux) os=linux ;; *) fail 'Use the Windows ZIP from GitHub Releases.' ;; esac
case "$(uname -m)" in x86_64|amd64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; *) fail 'Supported architectures: amd64 and arm64.' ;; esac
for tool in curl tar mktemp; do command -v "$tool" >/dev/null 2>&1 || fail "Required command not found: $tool"; done
if command -v sha256sum >/dev/null 2>&1; then hash_tool=sha256sum
elif command -v shasum >/dev/null 2>&1; then hash_tool=shasum
else fail 'SHA-256 verification requires sha256sum or shasum.'; fi

base=https://github.com/michaelmjhhhh/typeit/releases
if [ "$version" = latest ]; then base=$base/latest/download; else base=$base/download/$version; fi
asset=typeit_${os}_${arch}.tar.gz
tmp=$(mktemp -d)
staged=
trap 'rm -rf "$tmp"; if [ -n "$staged" ]; then rm -f "$staged"; fi' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
printf 'Downloading Typeit for %s/%s…\n' "$os" "$arch"
curl --proto '=https' --tlsv1.2 -fsSL --retry 3 "$base/$asset" -o "$tmp/$asset" || fail 'Download failed. Check the version and GitHub Releases.'
curl --proto '=https' --tlsv1.2 -fsSL --retry 3 "$base/SHA256SUMS" -o "$tmp/SHA256SUMS" || fail 'Could not download checksums.'
expected=$(awk -v file="$asset" '$2 == file { print $1 }' "$tmp/SHA256SUMS")
printf '%s\n' "$expected" | grep -Eq '^[a-fA-F0-9]{64}$' || fail 'Missing or invalid archive checksum.'
if [ "$hash_tool" = sha256sum ]; then actual=$(sha256sum "$tmp/$asset" | awk '{print $1}')
else actual=$(shasum -a 256 "$tmp/$asset" | awk '{print $1}'); fi
[ "$actual" = "$expected" ] || fail 'Checksum mismatch; nothing was installed.'
tar -xzf "$tmp/$asset" -C "$tmp" typeit LICENSE NOTICE THIRD_PARTY_NOTICES.md
[ -f "$tmp/typeit" ] && [ ! -L "$tmp/typeit" ] || fail 'Archive has no valid executable.'
install_dir=${TYPEIT_INSTALL_DIR:-"$HOME/.local/bin"}
licenses_dir=${XDG_DATA_HOME:-"$HOME/.local/share"}/typeit
mkdir -p "$install_dir" "$licenses_dir"
[ ! -d "$install_dir/typeit" ] || fail 'The install destination is a directory.'
cp "$tmp/LICENSE" "$tmp/NOTICE" "$tmp/THIRD_PARTY_NOTICES.md" "$licenses_dir/"
staged=$(mktemp "$install_dir/.typeit.XXXXXX")
cp "$tmp/typeit" "$staged"
chmod 755 "$staged"
mv -f "$staged" "$install_dir/typeit"
staged=
printf 'Installed %s\n' "$install_dir/typeit"
case ":$PATH:" in
    *":$install_dir:"*) printf 'Run: typeit\n' ;;
    *) printf 'Add the install directory to PATH, then run typeit:\n  export PATH="%s:$PATH"\n' "$install_dir" ;;
esac
