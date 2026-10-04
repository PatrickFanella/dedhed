import importlib.util
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import tempfile
import threading
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('public_runtime', Path(__file__).with_name('public-runtime.py'))
runtime = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runtime)


class PublicMaintenanceTests(unittest.TestCase):
    def response_server(self, statuses):
        seen = []
        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                assert self.headers['User-Agent'] == 'Dedhed-public-monitor/1.0'
                code = statuses[min(len(seen), len(statuses) - 1)]
                seen.append(code)
                self.send_response(code)
                self.send_header('content-range', 'bytes 0-4095/69657246')
                self.end_headers()
                self.wfile.write(b'failed response' if code != 200 else json.dumps({'ready': True}).encode())
            def log_message(self, *_):
                pass
        server = ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        self.addCleanup(server.server_close)
        self.addCleanup(server.shutdown)
        return f'http://127.0.0.1:{server.server_port}/', seen

    def test_transient_response_does_not_contaminate_retry(self):
        url, seen = self.response_server([503, 200])
        with patch.object(runtime.time, 'sleep'):
            _, headers, body = runtime.request(url)
            self.assertEqual(json.loads(body), {'ready': True})
            self.assertEqual(headers['content-range'], 'bytes 0-4095/69657246')
        self.assertEqual(seen, [503, 200])

    def test_access_rejection_remains_visible_without_retry(self):
        url, seen = self.response_server([403, 200])
        with self.assertRaisesRegex(RuntimeError, 'HTTP 403'):
            runtime.request(url)
        self.assertEqual(seen, [403])

    def test_capacity_exhaustion_preserves_checkpoint(self):
        with tempfile.TemporaryDirectory() as name:
            root = Path(name)
            evidence = root / 'checkpoint.dump'
            evidence.write_bytes(b'recovery evidence')
            with self.assertRaisesRegex(RuntimeError, 'no deletion performed'):
                runtime.capacity(root, 16, 0, 1)
            self.assertEqual(evidence.read_bytes(), b'recovery evidence')
