import unittest
from cmdbox.runtime.results import ExecutionResult


class TestExecutionResult(unittest.TestCase):

    def test_execution_result_creation(self):
        result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello\n", stderr=""
        )
        self.assertEqual(result.command, "echo hello")
        self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.stdout, "hello\n")
        self.assertEqual(result.stderr, "")

    def test_execution_result_immutability(self):
        result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello\n", stderr=""
        )
        with self.assertRaises(AttributeError):
            result.exit_code = 1
