"""Build a native, self-contained Typeit release archive."""
from pathlib import Path
import argparse
import os
import re
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parent.parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--version", required=True)
args = parser.parse_args()
if not re.fullmatch(r"v[0-9]+\.[0-9]+\.[0-9]+", args.version):
    parser.error("version must be vX.Y.Z")
version = args.version[1:]
goos, goarch = subprocess.check_output(["go", "env", "GOOS", "GOARCH"], text=True).split()
if (goos, goarch) not in {("darwin", "amd64"), ("darwin", "arm64"), ("linux", "amd64"), ("linux", "arm64"), ("windows", "amd64")}:
    parser.error(f"unsupported release platform: {goos}/{goarch}")
subprocess.run(["python3" if os.name != "nt" else "python", str(ROOT / "scripts/port_grammars.py"), "--ensure"], check=True)
env = {**os.environ, "CGO_ENABLED": "1"}
flags = f"-s -w -X github.com/michaelmjhhhh/typeit/internal/cli.Version={version}"
tags = "sqlite_omit_load_extension"
if goos == "linux":
    env["CC"] = env.get("CC") or "musl-gcc"
    if not shutil.which(env["CC"]):
        parser.error("Linux release builds require musl-gcc (musl-tools)")
    flags += " -linkmode external -extldflags=-static"
    tags += ",netgo,osusergo"
elif goos == "darwin":
    env.setdefault("MACOSX_DEPLOYMENT_TARGET", "14.0")
    minimum = env["MACOSX_DEPLOYMENT_TARGET"]
    if not re.fullmatch(r"[0-9]+\.[0-9]+", minimum):
        parser.error("invalid macOS deployment target")
    env["CGO_CFLAGS"] = env.get("CGO_CFLAGS", "-O2 -g") + f" -mmacosx-version-min={minimum}"
    env["CGO_LDFLAGS"] = env.get("CGO_LDFLAGS", "-O2 -g") + f" -mmacosx-version-min={minimum}"
else:
    env["CGO_LDFLAGS"] = env.get("CGO_LDFLAGS", "") + " -static-libgcc"
binary = ROOT / "bin" / ("typeit.exe" if goos == "windows" else "typeit")
binary.parent.mkdir(exist_ok=True)
subprocess.run(["go", "build", "-trimpath", "-tags", tags, "-ldflags", flags, "-o", str(binary), "./cmd/typeit"], cwd=ROOT, env=env, check=True)
runtime_env = dict(os.environ)
if goos == "windows":
    system_root = os.environ.get("SystemRoot", r"C:\Windows")
    runtime_env["PATH"] = os.pathsep.join([str(Path(system_root) / "System32"), system_root])
output = subprocess.check_output([str(binary), "--version"], env=runtime_env, text=True).strip()
if output != f"typeit {version}":
    raise RuntimeError(f"unexpected executable version: {output}")
(ROOT / "dist").mkdir(exist_ok=True)
archive = ROOT / "dist" / f"typeit_{goos}_{goarch}.{'zip' if goos == 'windows' else 'tar.gz'}"
with tempfile.TemporaryDirectory(prefix="typeit-package-") as temporary:
    stage = Path(temporary)
    for src in [binary, ROOT / "LICENSE", ROOT / "NOTICE", ROOT / "docs/THIRD_PARTY_NOTICES.md"]:
        shutil.copy2(src, stage / src.name)
    if goos == "windows":
        with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as out:
            for path in sorted(stage.iterdir()):
                out.write(path, path.name)
    else:
        def metadata(info):
            info.uid = info.gid = 0
            info.uname = info.gname = "root"
            info.mode = 0o755 if info.name == "typeit" else 0o644
            return info
        with tarfile.open(archive, "w:gz") as out:
            for path in sorted(stage.iterdir()):
                out.add(path, arcname=path.name, filter=metadata)
print(archive)
