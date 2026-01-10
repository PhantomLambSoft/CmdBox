import unittest
from unittest.mock import patch, MagicMock

from cmdbox.runtime.executor import Executor, RunContext
from cmdbox.runtime.results import ExecutionResult


class TestExecutor(unittest.TestCase):

    def setUp(self):
        self.executor = Executor()

    @patch("cmdbox.runtime.executor.subprocess.run")
    @patch("cmdbox.runtime.executor.build_shell_command")
    def test_run_success(self, mock_build, mock_run):
        # Setup
        command = "echo hello"
        popen_args = ["cmd.exe", "/c", "echo hello"]
        mock_build.return_value = popen_args

        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = "hello\n"
        mock_completed.stderr = ""
        mock_run.return_value = mock_completed

        # Execute
        result = self.executor.run(command)

        # Assert
        self.assertIsInstance(result, ExecutionResult)
        self.assertEqual(result.command, command)
        self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.stdout, "hello\n")
        self.assertEqual(result.stderr, "")

        mock_build.assert_called_once_with(command, preferred_shell=None)
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        self.assertEqual(args[0], popen_args)
        self.assertIsNone(kwargs["cwd"])
        self.assertFalse(kwargs["capture_output"])

    @patch("cmdbox.runtime.executor.subprocess.run")
    def test_run_with_context(self, mock_run):
        # Setup
        command = "echo hello"
        ctx = RunContext(cwd="/tmp", env={"VAR": "VAL"}, capture=True)

        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = "hello\n"
        mock_completed.stderr = ""
        mock_run.return_value = mock_completed

        # Execute
        self.executor.run(command, ctx)

        # Assert
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        self.assertEqual(kwargs["cwd"], "/tmp")
        self.assertTrue(kwargs["capture_output"])
        self.assertEqual(kwargs["env"]["VAR"], "VAL")

    @patch("cmdbox.runtime.executor.subprocess.run")
    def test_run_failure(self, mock_run):
        # Setup
        command = "false"
        mock_completed = MagicMock()
        mock_completed.returncode = 1
        mock_completed.stdout = ""
        mock_completed.stderr = "error"
        mock_run.return_value = mock_completed

        # Execute
        result = self.executor.run(command)

        # Assert
        self.assertEqual(result.exit_code, 1)
        self.assertEqual(result.stderr, "error")

    def test_run_context_defaults(self):
        ctx = RunContext()
        self.assertIsNone(ctx.cwd)
        self.assertIsNone(ctx.env)
        self.assertFalse(ctx.capture)
