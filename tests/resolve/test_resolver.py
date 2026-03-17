import unittest
from unittest.mock import MagicMock
from cmdbox.resolve.resolver import Resolver
from cmdbox.resolve.lookup import ResolverLookup
from cmdbox.resolve.type_defs import CommandRecord, VariableRecord, RefKind, TraceStep
from cmdbox.resolve.errors import (
    MaxDepthExceeded,
    UnknownReference,
    CycleDetectionError,
)


class TestResolver(unittest.TestCase):

    def setUp(self):
        self.mock_lookup = MagicMock(spec=ResolverLookup)
        self.resolver = Resolver(self.mock_lookup)

    def test_resolve_plain_text(self):
        result = self.resolver.resolve("hello world")
        self.assertEqual(result.text, "hello world")
        self.assertEqual(result.trace, [])

    def test_resolve_simple_variable(self):
        self.mock_lookup.get_variable.return_value = VariableRecord("name", "world")
        result = self.resolver.resolve("hello <name>")
        self.assertEqual(result.text, "hello world")
        self.assertEqual(len(result.trace), 1)
        self.assertEqual(result.trace[0], TraceStep(RefKind.VARIABLE, "name", "world"))

    def test_resolve_simple_command(self):
        self.mock_lookup.get_command.return_value = CommandRecord("greet", "echo hello")
        result = self.resolver.resolve("<cmd:greet>")
        self.assertEqual(result.text, "echo hello")
        self.assertEqual(len(result.trace), 1)
        self.assertEqual(
            result.trace[0], TraceStep(RefKind.COMMAND, "greet", "echo hello")
        )

    def test_resolve_explicit_variable(self):
        self.mock_lookup.get_variable.return_value = VariableRecord("name", "world")
        result = self.resolver.resolve("hello <var:name>")
        self.assertEqual(result.text, "hello world")

    def test_nested_resolution(self):
        def lookup_var(n):
            data = {"first": "John", "last": "Doe", "full": "<first> <last>"}
            if n in data:
                return VariableRecord(n, data[n])
            return None

        self.mock_lookup.get_variable.side_effect = lookup_var

        result = self.resolver.resolve("User: <full>")
        self.assertEqual(result.text, "User: John Doe")
        # Trace should be: first, last, full (last in first out of recursion, but appended to trace)
        # Actually _resolve_inner appends to trace AFTER recursion
        self.assertEqual(len(result.trace), 3)
        # Expected trace: [TraceStep(VAR, 'first', 'John'), TraceStep(VAR, 'last', 'Doe'), TraceStep(VAR, 'full', 'John Doe')]
        self.assertEqual(result.trace[0].key, "first")
        self.assertEqual(result.trace[1].key, "last")
        self.assertEqual(result.trace[2].key, "full")

    def test_variable_resolution_with_runtime_vars(self):
        self.mock_lookup.get_variable.return_value = None
        result = self.resolver.resolve(
            template="hello <name>", runtime_vars={"name": "Colonel Homer"}
        )
        self.assertEqual(result.text, "hello Colonel Homer")

    def test_runtime_vars_overrides_stored_vars(self):
        self.mock_lookup.get_variable.return_value = VariableRecord("name", "world")
        result = self.resolver.resolve(
            template="hello <name>", runtime_vars={"name": "Colonel Homer"}
        )
        self.assertEqual(result.text, "hello Colonel Homer")

    def test_runtime_vars_of_different_name_do_not_affect_resolution(self):
        self.mock_lookup.get_variable.return_value = VariableRecord("name", "world")
        result = self.resolver.resolve(
            template="hello <name>", runtime_vars={"tulip": "Colonel Homer"}
        )
        self.assertEqual(result.text, "hello world")

    def test_mix_of_stored_and_runtime_vars(self):
        self.mock_lookup.get_variable.return_value = VariableRecord(
            "name", "Colonel Homer"
        )
        result = self.resolver.resolve(
            template="hello <name>, have you met <name_two>?",
            runtime_vars={"name_two": "Lurleen"},
        )
        self.assertEqual(result.text, "hello Colonel Homer, have you met Lurleen?")

    def test_unknown_runtime_vars_do_not_affect_resolution(self):
        self.mock_lookup.get_variable.return_value = VariableRecord("name", "world")
        result = self.resolver.resolve(
            template="hello <name>", runtime_vars={"song": "Bagged Me A Homer"}
        )
        self.assertEqual(result.text, "hello world")

    def test_escape_characters(self):
        result = self.resolver.resolve("escaped \\<bracket\\>")
        self.assertEqual(result.text, "escaped <bracket>")

        result = self.resolver.resolve("double backslash \\\\")
        self.assertEqual(result.text, "double backslash \\")

    def test_unclosed_bracket(self):
        result = self.resolver.resolve("text <unclosed")
        self.assertEqual(result.text, "text <unclosed")

    def test_unknown_reference_non_strict(self):
        self.mock_lookup.get_variable.return_value = None
        result = self.resolver.resolve("<unknown>")
        self.assertEqual(result.text, "unknown")

    def test_unknown_reference_strict(self):
        self.resolver = Resolver(self.mock_lookup, strict=True)
        self.mock_lookup.get_variable.return_value = None
        with self.assertRaises(UnknownReference) as cm:
            self.resolver.resolve("<unknown>")
        self.assertEqual(cm.exception.kind, "variable")
        self.assertEqual(cm.exception.key, "unknown")

    def test_max_depth_exceeded(self):
        self.resolver = Resolver(self.mock_lookup, max_depth=2)
        self.mock_lookup.get_variable.side_effect = lambda n: VariableRecord(
            n, f"<{int(n)+1}>"
        )
        with self.assertRaises(MaxDepthExceeded):
            self.resolver.resolve("<0>")

    def test_cycle_detection(self):
        self.mock_lookup.get_variable.side_effect = lambda n: (
            {"A": "<B>", "B": "<A>"}.get(n)
            and VariableRecord(n, {"A": "<B>", "B": "<A>"}[n])
            or None
        )
        with self.assertRaises(CycleDetectionError) as cm:
            self.resolver.resolve("<A>")
        self.assertIn("var:A -> var:B -> var:A", str(cm.exception))

    def test_complex_token_inner(self):
        # < cmd : my_cmd > (with spaces)
        self.mock_lookup.get_command.return_value = CommandRecord("my_cmd", "done")
        result = self.resolver.resolve("< cmd : my_cmd >")
        self.assertEqual(result.text, "done")

    def test_colon_in_key(self):
        # <var:name:with:colons>
        self.mock_lookup.get_variable.return_value = VariableRecord(
            "name:with:colons", "val"
        )
        result = self.resolver.resolve("<var:name:with:colons>")
        self.assertEqual(result.text, "val")

    def test_empty_bracket(self):
        result = self.resolver.resolve("<>")
        self.assertEqual(result.text, "")
