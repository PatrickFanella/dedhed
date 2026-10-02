"""The app's token sheet must stay a verbatim copy of the design-system source."""
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


class DesignTokens(unittest.TestCase):
    def test_web_tokens_match_design_source(self):
        source = (ROOT / 'docs/design/tokens.css').read_text()
        applied = (ROOT / 'web/src/tokens.css').read_text()
        self.assertEqual(source, applied, 'web/src/tokens.css differs from docs/design/tokens.css; copy the source over it')


if __name__ == '__main__':
    unittest.main()
