"""Regenerate pinned native grammars without requiring the upstream Rust project."""
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path, PurePosixPath
import argparse
import hashlib
import io
import json
import re
import tarfile
import urllib.request

ROOT = Path(__file__).resolve().parent.parent
MANIFEST = json.loads((ROOT / "scripts/grammars.json").read_text())
STATE = ROOT / ".grammar-cache/state.json"


def fingerprint():
    return hashlib.sha256((ROOT / "scripts/grammars.json").read_bytes() + Path(__file__).read_bytes()).hexdigest()


def generated_files():
    return sorted(p for p in (ROOT / "internal/grammars").rglob("*")
                  if p.is_file() and p.suffix in {".c", ".cc", ".h"} and p.parent.parent.name != "grammars")


def ready():
    try:
        state = json.loads(STATE.read_text())
        return state["manifest"] == fingerprint() and bool(state["files"]) and all(
            (ROOT / name).is_file() and hashlib.sha256((ROOT / name).read_bytes()).hexdigest() == checksum
            for name, checksum in state["files"].items())
    except (OSError, ValueError, KeyError):
        return False


def fetch(entry):
    language, crate, version = entry["language"], entry["name"], entry["version"]
    target = ROOT / "internal/grammars" / language
    source = PurePosixPath(entry["parser"]).parent
    archive = ROOT / ".grammar-cache" / f"{crate}-{version}.crate"
    archive.parent.mkdir(exist_ok=True)
    if not archive.exists():
        url = f"https://static.crates.io/crates/{crate}/{crate}-{version}.crate"
        with urllib.request.urlopen(url, timeout=90) as response:
            data = response.read()
        if hashlib.sha256(data).hexdigest() != entry["checksum"]:
            raise ValueError(f"{crate}: checksum mismatch")
        archive.write_bytes(data)
    data = archive.read_bytes()
    if hashlib.sha256(data).hexdigest() != entry["checksum"]:
        raise ValueError(f"{crate}: checksum mismatch")
    with tarfile.open(fileobj=io.BytesIO(data)) as tar:
        prefix = f"{crate}-{version}/"
        for member in tar.getmembers():
            if not member.isfile() or not member.name.startswith(prefix):
                continue
            relative = PurePosixPath(member.name[len(prefix):])
            if relative.is_absolute() or ".." in relative.parts:
                raise ValueError(f"{crate}: invalid archive path")
            native = source in relative.parents and relative.suffix in {".c", ".cc", ".h"}
            common = relative.parts[0] == "common" and relative.suffix == ".h"
            license_file = re.match(r"^(LICENSE|COPYING|NOTICE)", relative.name) and "node_modules" not in relative.parts
            if native or common or license_file:
                out = target / relative
                out.parent.mkdir(parents=True, exist_ok=True)
                out.write_bytes(tar.extractfile(member).read())
    print(language, version, flush=True)
    return language


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ensure", action="store_true", help="reuse verified generated sources")
    args = parser.parse_args()
    if args.ensure and ready():
        return
    with ThreadPoolExecutor(max_workers=5) as pool:
        list(pool.map(fetch, MANIFEST))
    state = {"manifest": fingerprint(), "files": {
        str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest()
        for p in generated_files()
    }}
    temporary = STATE.with_suffix(".tmp")
    temporary.write_text(json.dumps(state, indent=2) + "\n")
    temporary.replace(STATE)


if __name__ == "__main__":
    main()
