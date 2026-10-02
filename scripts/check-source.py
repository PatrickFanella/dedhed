from pathlib import Path
import hashlib
import json
import re

root = Path(__file__).resolve().parent.parent
vendor = root / 'backend/third_party/arkhiv'
provenance = json.loads((vendor / 'provenance.json').read_text())
for name, info in provenance['files'].items():
    assert hashlib.sha256((vendor / 'internetarchive' / name).read_bytes()).hexdigest() == info['sha256'], name
for path in [root / 'README.md', root / 'ROADMAP.md', *root.joinpath('docs').rglob('*.md')]:
    text = path.read_text()
    assert all(line == line.rstrip() for line in text.splitlines()), path
    for target in re.findall(r'\]\(([^)]+)\)', text):
        if not target.startswith(('https://', 'http://', '#', '/')):
            assert (path.parent / target.split('#')[0]).is_file(), (path, target)
print('Source provenance and documentation links passed')
