"""Verify the independently reviewed frontend SDK receipt."""
import json
from pathlib import Path
import re

root = Path(__file__).resolve().parents[1]
receipt = json.loads((root / 'deploy/viewer-release.json').read_text())
package = json.loads((root / 'web/package.json').read_text())
assert receipt['package'] == '@subcult/archive-viewer'
assert package['dependencies'][receipt['package']] == receipt['version']
assert re.fullmatch(r'[0-9a-f]{40}', receipt['source_revision'])
assert re.fullmatch(r'[0-9a-f]{64}', receipt['package_sha256'])
assert receipt['compatible_core_receipt'] == 'core-release-0.3.0.json'
# The SDK needs at least the core API it was qualified against; later 0.x cores
# keep API version 1 and add fields, so the API pin may move ahead independently.
core = json.loads((root / 'deploy/core-release.json').read_text())
assert '1' in core['api_versions']
assert tuple(map(int, core['version'].split('.'))) >= (0, 3, 0) and core['version'].startswith('0.')
print('Independent frontend SDK pin verified:', receipt['version'], receipt['source_revision'])
