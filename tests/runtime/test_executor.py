import subprocess
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
        self.assertIsNone(ctx.shell)
        self.assertIsNone(ctx.timeout)
        self.assertFalse(ctx.emit)
        self.assertFalse(ctx.verbose)

    def test_is_multiline(self):
        self.assertFalse(self.executor.is_multiline("echo hello"))
        self.assertFalse(self.executor.is_multiline("echo hello\n"))
        self.assertFalse(self.executor.is_multiline("\necho hello"))
        self.assertFalse(self.executor.is_multiline("\necho hello\n"))
        self.assertTrue(self.executor.is_multiline("echo hello\necho world"))
        self.assertTrue(self.executor.is_multiline("echo hello\r\necho world"))
        self.assertTrue(self.executor.is_multiline("\necho hello\necho world\n"))
        self.assertFalse(self.executor.is_multiline(""))
        self.assertFalse(self.executor.is_multiline("\n"))

    @patch("sys.stdout")
    def test_emit_command(self, mock_stdout):
        with self.assertRaises(Exception) as cm:
            self.executor.emit_command("echo hello")

        from typer import Exit

        self.assertIsInstance(cm.exception, Exit)
        self.assertEqual(cm.exception.exit_code, 0)
        mock_stdout.write.assert_called_with("echo hello\n")

    @patch("cmdbox.runtime.executor.Executor.emit_command")
    def test_run_emit(self, mock_emit):
        ctx = RunContext(emit=True)
        result = self.executor.run("echo hello", ctx)
        self.assertIsNone(result)
        mock_emit.assert_called_once_with("echo hello")

    def test_script_suffix_for_shell(self):
        self.assertEqual(self.executor.script_suffix_for_shell("cmd.exe"), ".cmd")
        self.assertEqual(self.executor.script_suffix_for_shell("cmd"), ".cmd")
        self.assertEqual(self.executor.script_suffix_for_shell("powershell"), ".ps1")
        self.assertEqual(self.executor.script_suffix_for_shell("pwsh"), ".ps1")
        self.assertEqual(self.executor.script_suffix_for_shell("fish"), ".fish")
        self.assertEqual(self.executor.script_suffix_for_shell("bash"), ".sh")
        self.assertEqual(self.executor.script_suffix_for_shell("zsh"), ".sh")
        self.assertEqual(self.executor.script_suffix_for_shell(""), ".sh")

    def test_script_header_for_shell(self):
        self.assertEqual(
            self.executor.script_header_for_shell("bash"), "#!/usr/bin/env bash\n"
        )
        self.assertEqual(
            self.executor.script_header_for_shell("zsh"), "#!/usr/bin/env zsh\n"
        )
        self.assertEqual(
            self.executor.script_header_for_shell("fish"), "#!/usr/bin/env fish\n"
        )
        self.assertEqual(self.executor.script_header_for_shell("cmd"), "")
        self.assertEqual(self.executor.script_header_for_shell("powershell"), "")

    def test_build_script_exec_args(self):
        self.assertEqual(
            self.executor.build_script_exec_args("test.cmd", "cmd"),
            ["cmd.exe", "/d", "/s", "/c", "test.cmd"],
        )
        self.assertEqual(
            self.executor.build_script_exec_args("test.ps1", "powershell"),
            [
                "powershell",
                "-NoProfile",
                "-NonInteractive",
                "-ExecutionPolicy",
                "Bypass",
                "-File",
                "test.ps1",
            ],
        )
        self.assertEqual(
            self.executor.build_script_exec_args("test.ps1", "pwsh"),
            [
                "pwsh",
                "-NoProfile",
                "-NonInteractive",
                "-ExecutionPolicy",
                "Bypass",
                "-File",
                "test.ps1",
            ],
        )
        self.assertEqual(
            self.executor.build_script_exec_args("test.fish", "fish"),
            ["fish", "test.fish"],
        )
        self.assertEqual(
            self.executor.build_script_exec_args("test.sh", "zsh"), ["zsh", "test.sh"]
        )
        self.assertEqual(
            self.executor.build_script_exec_args("test.sh", "bash"), ["bash", "test.sh"]
        )
        self.assertEqual(
            self.executor.build_script_exec_args("test.sh", ""), ["sh", "test.sh"]
        )

    @patch("cmdbox.runtime.executor.subprocess.run")
    @patch("os.remove")
    @patch("tempfile.NamedTemporaryFile")
    def test_run_multiline_as_script(self, mock_temp, mock_remove, mock_run):
        # Setup
        command = "echo hello\necho world"
        ctx = RunContext(cwd="/tmp", capture=True)
        env = {"VAR": "VAL"}

        mock_file = MagicMock()
        mock_file.name = "temp_script.sh"
        mock_temp.return_value.__enter__.return_value = mock_file

        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = "hello\nworld\n"
        mock_completed.stderr = ""
        mock_run.return_value = mock_completed

        # Execute
        result = self.executor.run_multiline_as_script(command, ctx, env)

        # Assert
        self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.stdout, "hello\nworld\n")
        mock_temp.assert_called_once()
        mock_file.write.assert_any_call("echo hello\necho world\n")
        mock_run.assert_called_once()
        mock_remove.assert_called_once_with("temp_script.sh")

    @patch("cmdbox.runtime.executor.subprocess.run")
    @patch("os.remove")
    @patch("tempfile.NamedTemporaryFile")
    def test_run_multiline_as_script_cleanup_on_failure(
        self, mock_temp, mock_remove, mock_run
    ):
        # Setup
        command = "echo hello\necho world"
        ctx = RunContext()
        env = {}

        mock_file = MagicMock()
        mock_file.name = "temp_script.sh"
        mock_temp.return_value.__enter__.return_value = mock_file

        mock_run.side_effect = RuntimeError("Execution failed")

        # Execute & Assert
        with self.assertRaises(RuntimeError):
            self.executor.run_multiline_as_script(command, ctx, env)

        mock_remove.assert_called_once_with("temp_script.sh")

    @patch("cmdbox.runtime.executor.Executor.run_multiline_as_script")
    def test_run_dispatch_multiline(self, mock_run_multiline):
        command = "echo hello\necho world"
        self.executor.run(command)
        mock_run_multiline.assert_called_once()

    @patch("cmdbox.runtime.executor.subprocess.run")
    def test_without_timeout_uses_subprocess_run(self, mock_run):
        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = "hello\n"
        mock_completed.stderr = ""
        mock_run.return_value = mock_completed

        result = self.executor.execute_command(
            "echo hello",
            popen_args=["echo", "hello"],
            cwd=None,
            env={},
            capture_output=False,
            timeout=None,
        )

        mock_run.assert_called_once()
        self.assertIsInstance(result, ExecutionResult)
        self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.stdout, "hello\n")

    @patch("cmdbox.runtime.executor.Executor._run_subprocess")
    def test_with_timeout_uses_run_subprocess(self, mock_run_subprocess):
        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = "hello\n"
        mock_completed.stderr = ""
        mock_run_subprocess.return_value = mock_completed

        result = self.executor.execute_command(
            "echo hello",
            popen_args=["echo", "hello"],
            timeout=5,
        )

        mock_run_subprocess.assert_called_once()
        self.assertEqual(result.exit_code, 0)

    @patch("cmdbox.runtime.executor.subprocess.run")
    @patch("cmdbox.runtime.executor.Executor._run_subprocess")
    def test_without_timeout_does_not_call_run_subprocess(
        self, mock_run_subprocess, mock_run
    ):
        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = ""
        mock_completed.stderr = ""
        mock_run.return_value = mock_completed

        self.executor.execute_command(
            "echo hello",
            popen_args=["echo", "hello"],
            timeout=None,
        )

        mock_run_subprocess.assert_not_called()

    @patch("cmdbox.runtime.executor.Executor._run_subprocess")
    def test_timeout_expired_returns_exit_code_124(self, mock_run_subprocess):
        mock_run_subprocess.side_effect = subprocess.TimeoutExpired(
            cmd="echo hello", timeout=5
        )

        result = self.executor.execute_command(
            "echo hello",
            popen_args=["echo", "hello"],
            timeout=5,
        )

        self.assertEqual(result.exit_code, 124)
        self.assertEqual(result.stdout, "")

    @patch("cmdbox.runtime.executor.Executor._run_subprocess")
    def test_timeout_expired_stderr_contains_timeout_value(self, mock_run_subprocess):
        mock_run_subprocess.side_effect = subprocess.TimeoutExpired(
            cmd="echo hello", timeout=10
        )

        result = self.executor.execute_command(
            "echo hello",
            popen_args=["echo", "hello"],
            timeout=10,
        )

        self.assertIn("10", result.stderr)

    @patch("cmdbox.runtime.executor.subprocess.run")
    def test_non_zero_exit_code_returned_correctly(self, mock_run):
        mock_completed = MagicMock()
        mock_completed.returncode = 1
        mock_completed.stdout = ""
        mock_completed.stderr = "error message"
        mock_run.return_value = mock_completed

        result = self.executor.execute_command(
            "false",
            popen_args=["false"],
            timeout=None,
        )

        self.assertEqual(result.exit_code, 1)
        self.assertEqual(result.stderr, "error message")

    @patch("cmdbox.runtime.executor.subprocess.run")
    def test_cwd_and_capture_passed_to_subprocess_run(self, mock_run):
        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = ""
        mock_completed.stderr = ""
        mock_run.return_value = mock_completed

        self.executor.execute_command(
            "echo hello",
            popen_args=["echo", "hello"],
            cwd="/some/path",
            env={"VAR": "val"},
            capture_output=True,
            timeout=None,
        )

        args, kwargs = mock_run.call_args
        self.assertEqual(kwargs["cwd"], "/some/path")
        self.assertTrue(kwargs["capture_output"])
        self.assertEqual(kwargs["env"]["VAR"], "val")

    @patch("cmdbox.runtime.executor.subprocess.Popen")
    def test_success_returns_completed_process_with_correct_values(self, mock_popen):
        mock_proc = MagicMock()
        mock_proc.returncode = 0
        mock_proc.communicate.return_value = ("stdout output", "stderr output")
        mock_popen.return_value.__enter__.return_value = mock_proc

        result = self.executor._run_subprocess(
            ["echo", "hello"],
            cwd=None,
            env={},
            capture_output=True,
            timeout=5,
        )

        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "stdout output")
        self.assertEqual(result.stderr, "stderr output")

    @patch("cmdbox.runtime.executor.subprocess.Popen")
    def test_communicate_called_with_correct_timeout(self, mock_popen):
        mock_proc = MagicMock()
        mock_proc.returncode = 0
        mock_proc.communicate.return_value = ("", "")
        mock_popen.return_value.__enter__.return_value = mock_proc

        self.executor._run_subprocess(
            ["echo", "hello"],
            cwd=None,
            env={},
            capture_output=False,
            timeout=7,
        )

        mock_proc.communicate.assert_called_once_with(timeout=7)

    @patch("cmdbox.runtime.executor.Executor._kill_process_tree")
    @patch("cmdbox.runtime.executor.subprocess.Popen")
    def test_timeout_expired_kills_process_tree_and_reraises(
        self, mock_popen, mock_kill
    ):
        mock_proc = MagicMock()
        mock_proc.communicate.side_effect = subprocess.TimeoutExpired(
            cmd="ping", timeout=2
        )
        mock_popen.return_value.__enter__.return_value = mock_proc

        with self.assertRaises(subprocess.TimeoutExpired):
            self.executor._run_subprocess(
                ["ping", "127.0.0.1"],
                cwd=None,
                env={},
                capture_output=False,
                timeout=2,
            )

        mock_kill.assert_called_once_with(mock_proc)
        mock_proc.wait.assert_called_once()

    @patch("cmdbox.runtime.executor.Executor._kill_process_tree")
    @patch("cmdbox.runtime.executor.subprocess.Popen")
    def test_keyboard_interrupt_kills_process_tree_and_reraises(
        self, mock_popen, mock_kill
    ):
        mock_proc = MagicMock()
        mock_proc.communicate.side_effect = KeyboardInterrupt()
        mock_popen.return_value.__enter__.return_value = mock_proc

        with self.assertRaises(KeyboardInterrupt):
            self.executor._run_subprocess(
                ["ping", "127.0.0.1"],
                cwd=None,
                env={},
                capture_output=False,
                timeout=5,
            )

        mock_kill.assert_called_once_with(mock_proc)
        mock_proc.wait.assert_called_once()

    @patch("cmdbox.runtime.executor.Executor._kill_process_tree")
    @patch("cmdbox.runtime.executor.subprocess.Popen")
    def test_success_does_not_call_kill_process_tree(self, mock_popen, mock_kill):
        mock_proc = MagicMock()
        mock_proc.returncode = 0
        mock_proc.communicate.return_value = ("", "")
        mock_popen.return_value.__enter__.return_value = mock_proc

        self.executor._run_subprocess(
            ["echo", "hello"],
            cwd=None,
            env={},
            capture_output=False,
            timeout=5,
        )

        mock_kill.assert_not_called()

    @patch("cmdbox.runtime.executor.subprocess.run")
    @patch("cmdbox.runtime.executor.sys")
    def test_windows_calls_taskkill_with_correct_args(self, mock_sys, mock_run):
        mock_sys.platform = "win32"
        mock_proc = MagicMock()
        mock_proc.pid = 12345

        Executor._kill_process_tree(mock_proc)

        mock_run.assert_called_once_with(
            ["taskkill", "/F", "/T", "/PID", "12345"],
            capture_output=True,
        )

    @patch("cmdbox.runtime.executor.Executor._run_subprocess")
    @patch("os.remove")
    @patch("tempfile.NamedTemporaryFile")
    def test_with_timeout_uses_run_subprocess(
        self, mock_temp, mock_remove, mock_run_subprocess
    ):
        command = "echo hello\necho world"
        ctx = RunContext(capture=True, timeout=5)
        env = {"VAR": "VAL"}

        mock_file = MagicMock()
        mock_file.name = "temp_script.sh"
        mock_temp.return_value.__enter__.return_value = mock_file

        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = "hello\nworld\n"
        mock_completed.stderr = ""
        mock_run_subprocess.return_value = mock_completed

        result = self.executor.run_multiline_as_script(command, ctx, env)

        mock_run_subprocess.assert_called_once()
        self.assertEqual(result.exit_code, 0)

    @patch("cmdbox.runtime.executor.subprocess.run")
    @patch("cmdbox.runtime.executor.Executor._run_subprocess")
    @patch("os.remove")
    @patch("tempfile.NamedTemporaryFile")
    def test_without_timeout_does_not_use_run_subprocess(
        self, mock_temp, mock_remove, mock_run_subprocess, mock_run
    ):
        command = "echo hello\necho world"
        ctx = RunContext(capture=True)
        env = {}

        mock_file = MagicMock()
        mock_file.name = "temp_script.sh"
        mock_temp.return_value.__enter__.return_value = mock_file

        mock_completed = MagicMock()
        mock_completed.returncode = 0
        mock_completed.stdout = ""
        mock_completed.stderr = ""
        mock_run.return_value = mock_completed

        self.executor.run_multiline_as_script(command, ctx, env)

        mock_run_subprocess.assert_not_called()

    @patch("cmdbox.runtime.executor.Executor._run_subprocess")
    @patch("os.remove")
    @patch("tempfile.NamedTemporaryFile")
    def test_timeout_expired_returns_exit_code_124_and_cleans_up_temp_file(
        self, mock_temp, mock_remove, mock_run_subprocess
    ):
        command = "echo hello\necho world"
        ctx = RunContext(timeout=2)
        env = {}

        mock_file = MagicMock()
        mock_file.name = "temp_script.sh"
        mock_temp.return_value.__enter__.return_value = mock_file

        mock_run_subprocess.side_effect = subprocess.TimeoutExpired(
            cmd="temp_script.sh", timeout=2
        )

        result = self.executor.run_multiline_as_script(command, ctx, env)

        self.assertEqual(result.exit_code, 124)
        mock_remove.assert_called_once_with("temp_script.sh")
