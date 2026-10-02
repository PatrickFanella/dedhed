"""Observe a real-time ffmpeg delivery run without claiming browser playback."""
import argparse
from datetime import datetime, timezone
import ipaddress
import json
from pathlib import Path
import time
from urllib.parse import urlparse
from urllib.request import urlopen


def loopback_url(value):
    url = urlparse(value)
    if url.scheme != 'http' or url.username or url.password:
        raise ValueError('Observer requires an HTTP loopback API without credentials')
    if url.hostname != 'localhost' and not ipaddress.ip_address(url.hostname).is_loopback:
        raise ValueError('Observer requires a loopback API')
    return value.rstrip('/')


def progress(path):
    result = {}
    for line in path.read_text().splitlines():
        if '=' in line:
            key, value = line.split('=', 1)
            result[key] = value
    return result


def observe(api, log, errors, expected, output):
    if output.exists():
        raise ValueError('Observation output must be new')
    report = {'proof': 'real-time HTTP decode and API observation',
              'expected_seconds': expected, 'samples': [], 'complete': False}
    began = time.monotonic()
    while True:
        sample = {'at': datetime.now(timezone.utc).isoformat()}
        try:
            with urlopen(api + '/api/station', timeout=5) as response:
                sample['station'] = json.load(response)
            with urlopen(api + '/api/ready', timeout=5) as response:
                sample['ready_status'] = response.status
        except Exception as error:
            sample['error'] = str(error)
        state = progress(log)
        sample['decoded_seconds'] = int(state.get('out_time_us', '0')) / 1e6
        report['samples'].append(sample)
        report['decoder_errors'] = errors.read_text()
        ended = state.get('progress') == 'end'
        report['complete'] = ended and abs(sample['decoded_seconds'] - expected) < 0.1
        report['observation_seconds'] = time.monotonic() - began
        report['api_errors'] = sum('error' in s for s in report['samples'])
        output.write_text(json.dumps(report, indent=2) + '\n')
        if ended:
            return report['complete'] and not report['decoder_errors'] and report['api_errors'] == 0
        if time.monotonic() - began > expected + 120:
            return False
        time.sleep(15)


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--api', type=loopback_url, required=True)
    parser.add_argument('--progress', type=Path, required=True)
    parser.add_argument('--errors', type=Path, required=True)
    parser.add_argument('--expected-duration', type=float, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if not 0 < args.expected_duration <= 6 * 3600:
        parser.error('Expected duration must be positive and at most six hours')
    raise SystemExit(0 if observe(args.api, args.progress, args.errors,
                                 args.expected_duration, args.output) else 1)
