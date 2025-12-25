import unittest
from cmdbox.resolve.errors import (
    UnknownReference,
    MaxDepthExceeded,
    CycleDetectionError,
)


class TestErrors(unittest.TestCase):
    def test_unknown_reference(self):
        err = UnknownReference("command", "foo")
        self.assertEqual(str(err), "Unknown command: foo")
        self.assertEqual(err.kind, "command")
        self.assertEqual(err.key, "foo")

    def test_max_depth_exceeded(self):
        err = MaxDepthExceeded(10)
        self.assertEqual(str(err), "Maximum resolution depth exceeded: (10)")
        self.assertEqual(err.max_depth, 10)

    def test_cycle_detection_error(self):
        err = CycleDetectionError(["A", "B", "A"])
        self.assertEqual(str(err), "Cycle detected: A -> B -> A")
        self.assertEqual(err.path, ["A", "B", "A"])
