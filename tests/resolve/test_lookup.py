import unittest
from unittest.mock import MagicMock
from cmdbox.resolve.lookup import MemoizedLookup, ResolverLookup
from cmdbox.resolve.type_defs import CommandRecord, VariableRecord


class TestLookup(unittest.TestCase):

    def setUp(self):
        self.mock_inner = MagicMock(spec=ResolverLookup)
        self.memoized = MemoizedLookup(self.mock_inner)

    def test_get_command_caches(self):
        cmd = CommandRecord("foo", "echo bar")
        self.mock_inner.get_command.return_value = cmd

        # First call
        res1 = self.memoized.get_command("foo")
        self.assertEqual(res1, cmd)
        self.mock_inner.get_command.assert_called_once_with("foo")

        # Second call
        res2 = self.memoized.get_command("foo")
        self.assertEqual(res2, cmd)
        self.mock_inner.get_command.assert_called_once()

    def test_get_variable_caches(self):
        var = VariableRecord("var", "val")
        self.mock_inner.get_variable.return_value = var

        # First call
        res1 = self.memoized.get_variable("var")
        self.assertEqual(res1, var)
        self.mock_inner.get_variable.assert_called_once_with("var")

        # Second call
        res2 = self.memoized.get_variable("var")
        self.assertEqual(res2, var)
        self.mock_inner.get_variable.assert_called_once()

    def test_clear(self):
        self.mock_inner.get_command.return_value = CommandRecord("foo", "bar")
        self.memoized.get_command("foo")
        self.memoized.clear()

        self.memoized.get_command("foo")
        self.assertEqual(self.mock_inner.get_command.call_count, 2)
