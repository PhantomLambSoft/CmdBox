import unittest
from cmdbox.resolve.type_defs import (
    CommandRecord,
    VariableRecord,
    RefKind,
    TraceStep,
    ResolveResult,
)


class TestTypes(unittest.TestCase):

    def test_command_record(self):
        rec = CommandRecord(alias="f", template="t")
        self.assertEqual(rec.alias, "f")
        self.assertEqual(rec.template, "t")

    def test_variable_record(self):
        rec = VariableRecord(name="n", value="v")
        self.assertEqual(rec.name, "n")
        self.assertEqual(rec.value, "v")

    def test_ref_kind(self):
        self.assertEqual(RefKind.COMMAND, "command")
        self.assertEqual(RefKind.VARIABLE, "variable")

    def test_trace_step(self):
        step = TraceStep(kind=RefKind.COMMAND, key="k", expanded_to="e")
        self.assertEqual(step.kind, RefKind.COMMAND)
        self.assertEqual(step.key, "k")
        self.assertEqual(step.expanded_to, "e")

    def test_resolve_result(self):
        res = ResolveResult(text="txt", trace=[])
        self.assertEqual(res.text, "txt")
        self.assertEqual(res.trace, [])
