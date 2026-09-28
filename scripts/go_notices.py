"""Collect notices for the Go modules and pinned native grammars."""
from pathlib import Path
import json
import subprocess

ROOT = Path(__file__).resolve().parent.parent
raw = subprocess.check_output(["go", "list", "-m", "-json", "all"], cwd=ROOT, text=True)
decoder = json.JSONDecoder()
modules = []
while raw.strip():
    raw = raw.lstrip()
    module, end = decoder.raw_decode(raw)
    modules.append(module)
    raw = raw[end:]

text = ["# Third-party notices", "", "Typeit is MIT-licensed. Its dependencies retain their own licenses.", ""]

def add_notice(title, files):
    if not files:
        raise ValueError(f"No license files found for {title}")
    text.extend([f"## {title}", ""])
    for path in files:
        text.extend([f"### {path.name}", "", "```text", path.read_text(errors="replace").strip(), "```", ""])


def licenses(directory):
    return sorted(p for p in directory.iterdir() if p.is_file() and p.name.upper().startswith(("LICENSE", "COPYING", "NOTICE")))


goroot = Path(subprocess.check_output(["go", "env", "GOROOT"], text=True).strip())
go_license = next((p for p in [goroot / "LICENSE", goroot.parent / "LICENSE"] if p.is_file()), None)
if go_license is None:
    raise FileNotFoundError("Go toolchain license was not found")
add_notice("Go runtime and standard library", [go_license])
for module in modules:
    if module.get("Main"):
        continue
    directory = module.get("Dir")
    if not directory:
        continue
    files = licenses(Path(directory))
    if not files:
        files = sorted(p for p in Path(directory).glob("README*") if "license" in p.read_text(errors="replace").lower())
    add_notice(f'{module["Path"]} {module.get("Version", "")}', files)

add_notice("Adapted Charmbracelet Huh and Catppuccin theme palettes", licenses(ROOT / "assets/themes/licenses"))

for grammar in json.loads((ROOT / "scripts/grammars.json").read_text()):
    directory = ROOT / "internal/grammars" / grammar["language"]
    files = sorted(p for p in directory.rglob("*") if p.is_file()
                   and p.name.upper().startswith(("LICENSE", "COPYING")) and p.suffix != ".source")
    add_notice(f'{grammar["name"]} {grammar["version"]}', files)

(ROOT / "docs/THIRD_PARTY_NOTICES.md").write_text("\n".join(text))
