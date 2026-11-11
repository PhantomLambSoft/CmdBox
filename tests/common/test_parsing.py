import unittest

import pytest

from cmdbox.common.parsing import parse_byte_size


class TestParseByteSize(unittest.TestCase):

    def test_plain_bytes(self):
        self.assertEqual(parse_byte_size("10"), 10)
        self.assertEqual(parse_byte_size("0"), 0)
        self.assertEqual(parse_byte_size(4096), 4096)

    def test_units(self):
        self.assertEqual(parse_byte_size("10kb"), 10240)
        self.assertEqual(parse_byte_size("10KB"), 10240)
        self.assertEqual(parse_byte_size("1.5mb"), int(1.5 * 1024**2))
        self.assertEqual(parse_byte_size("2 gb"), 2 * 1024**3)
        self.assertEqual(parse_byte_size("3b"), 3)

    def test_invalid_strings_raise(self):
        for bad in ("", "b", "abc", "10k", "1.5"):
            with self.assertRaises(ValueError, msg=bad):
                parse_byte_size(bad)

    def test_negative_sizes_raise(self):
        # A negative size is never valid: it would silently disable
        # size-based log rotation downstream.
        for bad in ("-5", "-1kb", "-0.5mb", -10):
            with self.assertRaises(ValueError, msg=repr(bad)):
                parse_byte_size(bad)
