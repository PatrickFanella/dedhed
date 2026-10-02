"""Acquire only an explicitly reviewed audience recording; preserve originals.

Run outside isolated previews. All payloads, receipts, and generated files stay
in the selected private data directory, never in the repository.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request
import wave


class SafeRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        parsed = urllib.parse.urlsplit(newurl)
        if parsed.scheme != 'https' or parsed.username or parsed.port not in (None, 443) or not (parsed.hostname == 'archive.org' or (parsed.hostname or '').endswith('.us.archive.org') or re.fullmatch(r'dn[0-9]+\.ca\.archive\.org', parsed.hostname or '')):
            raise ValueError('Source redirect rejected')
        return super().redirect_request(req, fp, code, msg, headers, newurl)


def digest(path, algorithm='sha256'):
    h = hashlib.new(algorithm)
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def acquire(recording, review, root, total_cap=1024 * 1024 * 1024, file_cap=250 * 1024 * 1024):
    if review.get('kind') != 'audience' or not review.get('acquisition') or not review.get('private_only') or review.get('relay'):
        raise ValueError('Explicit private audience acquisition decision required')
    item = recording['id']
    if not re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9_.-]{0,200}', item):
        raise ValueError('Invalid source identifier')
    tracks = recording['tracks']
    if not 1 <= len(tracks) <= 50:
        raise ValueError('Track count exceeds pilot bounds')
    if sum(t['file']['bytes'] for t in tracks) > total_cap:
        raise ValueError('Planned bytes exceed cap')
    root.mkdir(parents=True, exist_ok=True, mode=0o700)
    originals = root / 'originals'; originals.mkdir(exist_ok=True, mode=0o700)
    media = root / 'media'; media.mkdir(exist_ok=True, mode=0o700)
    opener = urllib.request.build_opener(SafeRedirect())
    receipt = {'item': item, 'review': review, 'total_cap': total_cap, 'file_cap': file_cap, 'concurrency': 1, 'files': [], 'errors': [], 'started_at': time.time()}
    receipt_path = root / (item + '.acquisition.json')
    def save():
        receipt_path.write_text(json.dumps(receipt, indent=2) + '\n'); receipt_path.chmod(0o600)
    save()
    total = 0; paths = []
    try:
        for t in tracks:
            f = t['file']; name = f['name']
            if Path(name).name != name or not re.fullmatch(r'[A-Za-z0-9_.-]+', name) or f['format'] != 'VBR MP3' or not 0 < f['bytes'] <= file_cap:
                raise ValueError('Unsafe or ineligible file: ' + name)
            md5 = f.get('md5', '')
            if not re.fullmatch(r'[a-fA-F0-9]{32}', md5):
                raise ValueError('Source checksum required for acquisition')
            known = list(originals.glob('*.mp3'))
            target = next((p for p in known if p.stat().st_size == f['bytes'] and digest(p, 'md5') == md5), None)
            if target is None:
                part = root / (item + '-' + name + '.part')
                url = 'https://archive.org/download/' + urllib.parse.quote(item, safe='') + '/' + urllib.parse.quote(name, safe='')
                for attempt in range(4):
                    start = part.stat().st_size if part.exists() else 0
                    req = urllib.request.Request(url, headers={'User-Agent': 'Dedhed/0.1 (bounded private music pilot)', 'Accept-Encoding': 'identity', **({'Range': f'bytes={start}-'} if start else {})})
                    try:
                        with opener.open(req, timeout=40) as response:
                            resume = start > 0 and response.status == 206
                            if resume and not response.headers.get('Content-Range', '').startswith(f'bytes {start}-'):
                                raise ValueError('Incorrect resume response')
                            if response.status not in (200, 206):
                                raise ValueError('Unexpected source status')
                            if not resume:
                                start = 0
                            with part.open('ab' if resume else 'wb') as dest:
                                size = start
                                while block := response.read(256 * 1024):
                                    size += len(block)
                                    if size > f['bytes'] or size > file_cap or total + size > total_cap:
                                        raise ValueError('Transfer exceeds approved bounds')
                                    dest.write(block)
                        if part.stat().st_size != f['bytes'] or digest(part, 'md5') != md5:
                            # Do not resume from bytes that failed source identity verification.
                            part.unlink(missing_ok=True)
                            raise ValueError('Source size/checksum mismatch')
                        target = originals / (digest(part) + '.mp3'); os.replace(part, target); target.chmod(0o600)
                        break
                    except (urllib.error.URLError, TimeoutError, OSError) as error:
                        if attempt == 3:
                            raise
                        if isinstance(error, urllib.error.HTTPError) and error.code not in (429, 500, 502, 503, 504):
                            raise
                        delay = min(2 ** attempt, 8)
                        if isinstance(error, urllib.error.HTTPError) and error.headers.get('Retry-After', '').isdigit():
                            delay = min(int(error.headers['Retry-After']), 60)
                        time.sleep(delay)
                if target is None:
                    raise ValueError('Transfer failed')
            total += target.stat().st_size
            if total > total_cap:
                raise ValueError('Actual bytes exceed cap')
            paths.append(target)
            receipt['files'].append({'source_name': name, 'bytes': target.stat().st_size, 'md5': md5, 'sha256': digest(target), 'path': str(target)})
            save(); print('Verified', name, flush=True)
        # Decode each track to the same PCM format, deriving boundaries from sample
        # counts rather than rounded source metadata. Concatenation adds no fades.
        with tempfile.TemporaryDirectory(prefix='dedhed-decode-', dir=root) as temporary:
            temp = Path(temporary); wavs = []; offset = 0.0
            for i, (track, original) in enumerate(zip(tracks, paths, strict=True)):
                wav = temp / f'{i:03d}.wav'
                subprocess.run(['ffmpeg', '-v', 'error', '-nostdin', '-i', str(original), '-ar', '48000', '-ac', '2', '-c:a', 'pcm_s16le', str(wav)], check=True)
                with wave.open(str(wav)) as audio:
                    duration = audio.getnframes() / audio.getframerate()
                track['start'] = offset; track['duration'] = duration; offset += duration; wavs.append(wav)
            listing = temp / 'concat.txt'; listing.write_text(''.join("file '" + p.name + "'\n" for p in wavs))
            generated = temp / 'continuous.m4a'
            subprocess.run(['ffmpeg', '-v', 'error', '-nostdin', '-f', 'concat', '-safe', '1', '-i', str(listing), '-c:a', 'aac', '-b:a', '160k', '-movflags', '+faststart', str(generated)], check=True)
            probe = json.loads(subprocess.check_output(['ffprobe', '-v', 'error', '-show_format', '-show_streams', '-of', 'json', str(generated)], text=True))
            if abs(float(probe['format']['duration']) - offset) > 0.1:
                raise ValueError('Derivative duration differs from decoded source samples')
            sha = digest(generated); target = media / (sha + '.m4a'); os.replace(generated, target); target.chmod(0o600)
            recording['duration'] = offset
            receipt.update({'derivative': {'name': target.name, 'sha256': sha, 'bytes': target.stat().st_size, 'duration': offset, 'probe': probe}, 'decoded_track_durations': [t['duration'] for t in tracks], 'ffmpeg_version': subprocess.check_output(['ffmpeg', '-version'], text=True).splitlines()[0], 'completed_at': time.time(), 'actual_source_bytes': total})
            (root / (item + '.playback.json')).write_text(json.dumps(recording, indent=2) + '\n')
            save()
    except Exception as e:
        receipt['errors'].append(str(e)); save(); raise
    return receipt


if __name__ == '__main__':
    parser = argparse.ArgumentParser(); parser.add_argument('--recording', type=Path, required=True); parser.add_argument('--review', type=Path, required=True); parser.add_argument('--root', type=Path, required=True)
    args = parser.parse_args()
    acquire(json.loads(args.recording.read_text()), json.loads(args.review.read_text()), args.root)
