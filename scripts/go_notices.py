from pathlib import Path
import json, subprocess
root=Path(__file__).resolve().parent.parent
raw=subprocess.check_output(['go','list','-m','-json','all'],cwd=root,text=True)
decoder=json.JSONDecoder();modules=[]
while raw.strip():
 raw=raw.lstrip();obj,end=decoder.raw_decode(raw);modules.append(obj);raw=raw[end:]
text=['# Go dependency notices','','The original GitType MIT license is retained in `LICENSE`. Native grammar licenses',
      'are retained beside the parsers under `internal/grammars`. This file contains',
      'license notices from the Go modules used by this port.','']
for module in modules:
 if module.get('Main') or 'Dir' not in module:continue
 directory=Path(module['Dir']);files=sorted(p for p in directory.iterdir() if p.is_file() and (p.name.upper().startswith(('LICENSE','COPYING','NOTICE'))))
 if not files:continue
 text += [f'## {module["Path"]} {module.get("Version", "")}','']
 for path in files:
  text += [f'### {path.name}','','```text',path.read_text(errors='replace').strip(),'```','']
(root/'THIRD_PARTY_GO.md').write_text('\n'.join(text))
