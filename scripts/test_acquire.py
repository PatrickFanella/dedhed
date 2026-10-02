import importlib.util
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location('acquire', Path(__file__).with_name('acquire.py'))
acquire = importlib.util.module_from_spec(spec); spec.loader.exec_module(acquire)


class AcquisitionPolicy(unittest.TestCase):
    def test_rejects_uncertain_or_relay_sources_before_network(self):
        for review in [{'kind': 'soundboard', 'acquisition': True}, {'kind': 'audience', 'acquisition': True, 'private_only': True, 'relay': True}, {}]:
            with tempfile.TemporaryDirectory() as root, self.assertRaises(ValueError):
                acquire.acquire({'id': 'sample', 'tracks': []}, review, Path(root))

    def test_rejects_traversal_before_network(self):
        with tempfile.TemporaryDirectory() as root, self.assertRaises(ValueError):
            acquire.acquire({'id': '../sample', 'tracks': []}, {'kind': 'audience', 'acquisition': True, 'private_only': True}, Path(root))

    def test_rejects_excess_planned_bytes_before_network(self):
        with tempfile.TemporaryDirectory() as root, self.assertRaises(ValueError):
            acquire.acquire({'id': 'sample', 'tracks': [{'file': {'bytes': 2000}}]}, {'kind': 'audience', 'acquisition': True, 'private_only': True}, Path(root), total_cap=1000)


if __name__ == '__main__': unittest.main()
