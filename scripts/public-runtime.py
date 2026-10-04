"""Bounded maintenance for an explicitly configured public Dedhed pilot."""
import argparse
import fcntl
import hashlib
import json
from pathlib import Path
import shlex
import shutil
import subprocess
import time
from datetime import datetime, timezone
from urllib.request import Request, urlopen
from urllib.error import HTTPError, URLError


def run(args, **kwargs):
    return subprocess.run(args, check=True, timeout=600, **kwargs)


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def request(url, headers=None):
    for attempt in range(3):
        try:
            with urlopen(Request(url, headers={'User-Agent': 'Dedhed-public-monitor/1.0', **(headers or {})}), timeout=15) as r:
                return r.status, {key.lower(): value for key, value in r.headers.items()}, r.read(1024 * 1024)
        except HTTPError as e:
            if e.code not in (429, 500, 502, 503, 504) or attempt == 2:
                raise RuntimeError(f'{url}: HTTP {e.code}') from e
        except URLError as e:
            if attempt == 2:
                raise RuntimeError(f'{url}: transport {e.reason}') from e
        time.sleep(attempt + 1)


def monitor(config):
    base = config['monitor_base'].rstrip('/')
    def get(path):
        return json.loads(request(base + path)[2])
    health = get('/api/health')
    if health.get('audience') != 'public' or health.get('revision') != config['revision']:
        raise RuntimeError('public runtime identity differs from configured release')
    catalog = get('/api/recordings')
    expected = set(config['allowed_ids'])
    actual = {e['recording']['id'] for e in catalog['items']}
    if actual != expected or catalog['total'] != len(expected):
        raise RuntimeError('public recording scope differs from approved pilot')
    station = get('/api/station')
    offset = station['offset']
    if station['program']['recording_id'] not in expected or not 0 <= offset < station['program']['duration']:
        raise RuntimeError('station is outside the approved timeline')
    if abs(datetime.fromisoformat(station['server_time'].replace('Z', '+00:00')).timestamp() - time.time()) > 30:
        raise RuntimeError('station server time is stale')
    for item in sorted(expected):
        status, headers, body = request(base + '/api/audio/' + item, {'Range': 'bytes=0-4095'})
        if status != 206 or len(body) != 4096 or not headers.get('content-range', '').startswith('bytes 0-4095/'):
            raise RuntimeError(f'approved audio range unavailable: {item}')
    return {'recordings': len(expected), 'schedule_revision': station['revision'], 'offset': offset}


def capacity(root, cap, reserve, allowance):
    used = sum(f.stat().st_size for f in root.rglob('*') if f.is_file())
    if used + allowance > cap or shutil.disk_usage(root).free - allowance < reserve:
        raise RuntimeError('checkpoint capacity or free-space reserve exhausted; no deletion performed')


def backup(config, root):
    checkpoints = root / 'checkpoints'
    checkpoints.mkdir(mode=0o700, exist_ok=True)
    capacity(checkpoints, config['checkpoint_cap'], config['reserve'], config['dump_allowance'])
    stamp = datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    checkpoint = checkpoints / stamp
    checkpoint.mkdir(mode=0o700)
    part = checkpoint / 'database.dump.part'
    with part.open('wb') as f:
        run(['docker', 'exec', config['database_container'], 'pg_dump', '-U', 'dedhed_operator', '-d', 'dedhed', '-Fc'], stdout=f)
    if part.stat().st_size > config['dump_allowance']:
        raise RuntimeError('database dump exceeds configured allowance; incomplete checkpoint retained')
    dump = checkpoint / 'database.dump'
    part.rename(dump)
    media = []
    for name, expected in config['media'].items():
        path = root / 'media' / name
        if digest(path) != expected:
            raise RuntimeError(f'public media checksum mismatch: {name}')
        media.append({'name': name, 'sha256': expected, 'bytes': path.stat().st_size})
    manifest = {'observed_at': datetime.now(timezone.utc).isoformat(), 'revision': config['revision'],
                'images': config['images'], 'dump_sha256': digest(dump), 'dump_bytes': dump.stat().st_size,
                'media': media, 'restore_qualified': False}
    receipt = checkpoint / 'manifest.json'
    receipt.write_text(json.dumps(manifest, indent=2) + '\n')
    host, dest = config['backup_host'], config['backup_root']
    remote = dest + '/' + stamp
    run(['ssh', host, 'mkdir -p ' + shlex.quote(remote) + ' ' + shlex.quote(dest + '/media')])
    # Check destination capacity before transfer. Never prune existing recovery evidence.
    allowance = config['dump_allowance'] + sum(x['bytes'] for x in media)
    remote_script = ('import pathlib,shutil; p=pathlib.Path(' + repr(dest)
                     + '); used=sum(f.stat().st_size for f in p.rglob("*") if f.is_file()); assert used+'
                     + str(allowance) + '<=' + str(config['checkpoint_cap'])
                     + '; assert shutil.disk_usage(p).free-' + str(allowance) + '>' + str(config['reserve']))
    remote_check = 'python3 -c ' + shlex.quote(remote_script)
    run(['ssh', host, remote_check])
    for entry in media:
        name = entry['name']
        remote_file = dest + '/media/' + name
        probe = run(['ssh', host, 'if test -f ' + shlex.quote(remote_file) + '; then sha256sum ' + shlex.quote(remote_file) + '; fi'], stdout=subprocess.PIPE, text=True).stdout.split()
        if probe:
            if probe[0] != entry['sha256']:
                raise RuntimeError('existing independent media differs; evidence preserved')
        else:
            run(['scp', str(root / 'media' / name), host + ':' + remote_file + '.part'], stdout=subprocess.DEVNULL)
            found = subprocess.check_output(['ssh', host, 'sha256sum ' + shlex.quote(remote_file + '.part')], text=True, timeout=60).split()[0]
            if found != entry['sha256']:
                raise RuntimeError('independent media checksum mismatch')
            run(['ssh', host, 'mv ' + shlex.quote(remote_file + '.part') + ' ' + shlex.quote(remote_file)])
    run(['scp', str(dump), str(receipt), host + ':' + remote + '/'], stdout=subprocess.DEVNULL)
    found = subprocess.check_output(['ssh', host, 'sha256sum ' + shlex.quote(remote + '/database.dump')], text=True, timeout=60).split()[0]
    if found != manifest['dump_sha256']:
        raise RuntimeError('independent database checksum mismatch')
    return {'checkpoint': stamp, 'independent_host': host, 'copy_verified': True, 'restore_qualified': False}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('action', choices=['monitor', 'replay', 'backup'])
    parser.add_argument('--root', type=Path, required=True)
    args = parser.parse_args()
    root = args.root.resolve()
    config = json.loads((root / 'ops' / 'config.json').read_text())
    with (root / 'ops' / 'maintenance.lock').open('a') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        if args.action == 'monitor':
            result = monitor(config)
        elif args.action == 'replay':
            process = run(['docker', 'compose', '-f', str(root / 'compose.json'), 'run', '--rm', 'operator',
                           'public-schedule', '-ids', ','.join(config['allowed_ids']), '-hours', '168'],
                          stdout=subprocess.PIPE, text=True)
            result = json.loads(process.stdout)
        else:
            result = backup(config, root)
        print(json.dumps({'action': args.action, 'observed_at': datetime.now(timezone.utc).isoformat(), **result}))


if __name__ == '__main__':
    main()
