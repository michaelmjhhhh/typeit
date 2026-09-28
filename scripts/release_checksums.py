"""Create SHA256SUMS only when all five platform archives are present."""
from pathlib import Path
import hashlib
import shutil

ROOT = Path(__file__).resolve().parent.parent
DIST = ROOT / "dist"
expected = [f"typeit_{os}_{arch}.tar.gz" for os in ["darwin", "linux"] for arch in ["amd64", "arm64"]]
expected.append("typeit_windows_amd64.zip")
for name in expected:
    if not (DIST / name).is_file():
        raise SystemExit(f"missing release archive: {name}")
shutil.copy2(ROOT / "install.sh", DIST / "install.sh")
entries = sorted(expected + ["install.sh"])
(DIST / "SHA256SUMS").write_text("".join(f"{hashlib.sha256((DIST / name).read_bytes()).hexdigest()}  {name}\n" for name in entries))
