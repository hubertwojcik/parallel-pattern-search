import unittest

from gen_text import gen_text, parse_size
from gen_patterns import gen_patterns


class TestParseSize(unittest.TestCase):
    def test_kb(self):
        self.assertEqual(parse_size("1kb"), 1_000)

    def test_mb(self):
        self.assertEqual(parse_size("10mb"), 10_000_000)

    def test_gb(self):
        self.assertEqual(parse_size("1gb"), 1_000_000_000)

    def test_raw_int(self):
        self.assertEqual(parse_size("500"), 500)

    def test_uppercase(self):
        self.assertEqual(parse_size("5MB"), 5_000_000)


class TestGenText(unittest.TestCase):
    def test_returns_bytes(self):
        result = gen_text(1_000, seed=42)
        self.assertIsInstance(result, bytes)

    def test_approximate_size(self):
        result = gen_text(1_000, seed=42)
        self.assertGreaterEqual(len(result), 1_000)

    def test_reproducible(self):
        a = gen_text(1_000, seed=42)
        b = gen_text(1_000, seed=42)
        self.assertEqual(a, b)

    def test_different_seeds(self):
        a = gen_text(1_000, seed=1)
        b = gen_text(1_000, seed=2)
        self.assertNotEqual(a, b)

    def test_valid_utf8(self):
        result = gen_text(5_000, seed=42)
        result.decode("utf-8")


class TestGenPatterns(unittest.TestCase):
    def test_count(self):
        patterns = gen_patterns(10, seed=42)
        self.assertEqual(len(patterns), 10)

    def test_unique(self):
        patterns = gen_patterns(50, seed=42)
        self.assertEqual(len(patterns), len(set(patterns)))

    def test_non_empty(self):
        patterns = gen_patterns(10, seed=42)
        for p in patterns:
            self.assertGreater(len(p), 0)

    def test_reproducible(self):
        a = gen_patterns(20, seed=42)
        b = gen_patterns(20, seed=42)
        self.assertEqual(a, b)

    def test_different_seeds(self):
        a = gen_patterns(20, seed=1)
        b = gen_patterns(20, seed=2)
        self.assertNotEqual(a, b)


if __name__ == "__main__":
    unittest.main()
