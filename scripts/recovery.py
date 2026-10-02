"""Private pilot backup, hash audit, and restore into an empty destination."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024 * 1024), b''): h.update(block)
    return h.hexdigest()


def checked_container(name, restore=False):
    if not re.fullmatch(r'dedhed-(private-pilot-db|restore-[a-zA-Z0-9-]+)', name):
        raise ValueError('Only explicitly named Dedhed private/restore containers are supported')
    if restore and not name.startswith('dedhed-restore-'):
        raise ValueError('Restore requires a separate dedhed-restore-* container')


def backup(root, destination, container):
    checked_container(container)
    if destination.exists(): raise ValueError('Backup destination must be new')
    destination.mkdir(parents=True, mode=0o700)
    data = destination / 'data'; data.mkdir(mode=0o700)
    for name in ['media', 'originals', 'sources']:
        source = root / name
        if source.exists():
            if any(p.is_symlink() for p in source.rglob('*')): raise ValueError('Symlinks cannot enter a private backup')
            shutil.copytree(source, data / name)
    for p in root.glob('*.json'): shutil.copy2(p, data / p.name)
    dump = destination / 'catalog.dump'
    with dump.open('wb') as f:
        subprocess.run(['docker', 'exec', container, 'pg_dump', '-U', 'dedhed', '-d', 'dedhed', '-Fc'], stdout=f, check=True)
    dump.chmod(0o600)
    manifest = {str(p.relative_to(destination)): sha(p) for p in sorted(destination.rglob('*')) if p.is_file()}
    (destination / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    (destination / 'manifest.json').chmod(0o600)
    return manifest


def audit(destination):
    manifest = json.loads((destination / 'manifest.json').read_text())
    for name, expected in manifest.items():
        path = destination / name
        if not path.resolve().is_relative_to(destination.resolve()) or path.is_symlink() or sha(path) != expected:
            raise ValueError('Backup integrity failure: ' + name)
    return len(manifest)


def restore(backup_dir, root, container):
    checked_container(container, restore=True)
    count = audit(backup_dir)
    if root.exists(): raise ValueError('Restore storage must be a new directory')
    tables = subprocess.check_output(['docker', 'exec', container, 'psql', '-U', 'dedhed', '-d', 'dedhed', '-Atc', "SELECT count(*) FROM pg_tables WHERE schemaname='public'"], text=True)
    if int(tables.strip()) != 0: raise ValueError('Restore database must be empty')
    with (backup_dir / 'catalog.dump').open('rb') as f:
        subprocess.run(['docker', 'exec', '-i', container, 'pg_restore', '-U', 'dedhed', '-d', 'dedhed', '--exit-on-error', '--no-owner'], stdin=f, check=True)
    shutil.copytree(backup_dir / 'data', root)
    return count


if __name__ == '__main__':
    parser = argparse.ArgumentParser(); parser.add_argument('action', choices=['backup', 'audit', 'restore']); parser.add_argument('--root', type=Path); parser.add_argument('--backup', type=Path, required=True); parser.add_argument('--db-container')
    args = parser.parse_args()
    if args.action == 'backup': print('Backed up', len(backup(args.root, args.backup, args.db_container)), 'hashed files')
    elif args.action == 'audit': print('Verified', audit(args.backup), 'backup hashes')
    else: print('Restored', restore(args.backup, args.root, args.db_container), 'hashed files into a fresh destination')
