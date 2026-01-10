import unittest
from unittest.mock import patch
import os

from cmdbox.runtime.shell import build_shell_command


class TestShell(unittest.TestCase):

    @patch("sys.platform", "win32")
    @patch("cmdbox.runtime.shell.which")
    def test_build_shell_command_windows_default(self, mock_which):
        # Mock cmd.exe as available
        mock_which.side_effect = lambda x: x == "cmd.exe"
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["cmd.exe", "/C", "echo hello"])

    @patch("sys.platform", "win32")
    @patch("cmdbox.runtime.shell.which")
    @patch.dict(os.environ, {"CMDBOX_SHELL": "powershell"}, clear=True)
    def test_build_shell_command_windows_powershell_env(self, mock_which):
        mock_which.side_effect = lambda x: x == "powershell"
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["powershell", "-NoProfile", "-Command", "echo hello"])

    @patch("sys.platform", "win32")
    @patch("cmdbox.runtime.shell.which")
    @patch.dict(os.environ, {"CMDBOX_SHELL": "pwsh"}, clear=True)
    def test_build_shell_command_windows_pwsh_env(self, mock_which):
        mock_which.side_effect = lambda x: x == "pwsh"
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["pwsh", "-NoProfile", "-Command", "echo hello"])

    @patch("sys.platform", "win32")
    @patch("cmdbox.runtime.shell.which")
    def test_build_shell_command_windows_preferred_shell(self, mock_which):
        mock_which.side_effect = lambda x: x == "powershell"
        command = "echo hello"
        result = build_shell_command(command, preferred_shell="powershell")
        self.assertEqual(result, ["powershell", "-NoProfile", "-Command", "echo hello"])

    @patch("sys.platform", "linux")
    @patch("cmdbox.runtime.shell.which")
    @patch("cmdbox.runtime.shell.os.path.exists")
    @patch("cmdbox.runtime.shell.os.path.isabs")
    @patch.dict(os.environ, {"SHELL": "/bin/bash"}, clear=True)
    def test_build_shell_command_linux_bash_env(
        self, mock_isabs, mock_exists, mock_which
    ):
        mock_isabs.side_effect = lambda x: x.startswith("/")
        mock_exists.side_effect = lambda x: x == "/bin/bash"
        mock_which.return_value = None
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["/bin/bash", "-lc", "echo hello"])

    @patch("sys.platform", "linux")
    @patch("cmdbox.runtime.shell.which")
    @patch("cmdbox.runtime.shell.os.path.exists")
    @patch("cmdbox.runtime.shell.os.path.isabs")
    @patch.dict(os.environ, {}, clear=True)
    def test_build_shell_command_linux_default_fallback(
        self, mock_isabs, mock_exists, mock_which
    ):
        # /bin/sh is usually available, /bin/bash is not
        mock_isabs.side_effect = lambda x: x.startswith("/")
        mock_exists.side_effect = lambda x: x == "/bin/sh"
        mock_which.return_value = None
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["/bin/sh", "-lc", "echo hello"])

    @patch("sys.platform", "linux")
    @patch("cmdbox.runtime.shell.which")
    @patch("cmdbox.runtime.shell.os.path.exists")
    @patch("cmdbox.runtime.shell.os.path.isabs")
    def test_build_shell_command_linux_preferred_shell(
        self, mock_isabs, mock_exists, mock_which
    ):
        mock_isabs.side_effect = lambda x: x.startswith("/")
        mock_exists.return_value = False
        mock_which.side_effect = lambda x: x == "zsh"
        command = "echo hello"
        result = build_shell_command(command, preferred_shell="zsh")
        self.assertEqual(result, ["zsh", "-lc", "echo hello"])

    @patch("sys.platform", "linux")
    @patch("cmdbox.runtime.shell.which")
    @patch("cmdbox.runtime.shell.os.path.exists")
    @patch("cmdbox.runtime.shell.os.path.isabs")
    @patch.dict(os.environ, {}, clear=True)
    def test_build_shell_command_linux_no_shell_found(
        self, mock_isabs, mock_exists, mock_which
    ):
        mock_isabs.side_effect = lambda x: x.startswith("/")
        mock_exists.return_value = False
        mock_which.return_value = None
        with self.assertRaises(RuntimeError) as cm:
            build_shell_command("echo hello")
        self.assertEqual(str(cm.exception), "No usable shell found for this system")

    @patch("sys.platform", "win32")
    @patch("cmdbox.runtime.shell.which")
    def test_build_shell_command_windows_fallback_order(self, mock_which):
        # Simulate pwsh and powershell missing, but cmd.exe available
        mock_which.side_effect = lambda x: x == "cmd.exe"
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["cmd.exe", "/C", "echo hello"])

    @patch("sys.platform", "win32")
    @patch("cmdbox.runtime.shell.which")
    def test_build_shell_command_windows_unknown_preferred_shell(self, mock_which):
        # If preferred shell is not a known shell, it's treated as an executable
        mock_which.side_effect = lambda x: x == "custom_exe"
        command = "echo hello"
        result = build_shell_command(command, preferred_shell="custom_exe")
        self.assertEqual(result, ["custom_exe", "echo hello"])

    @patch("sys.platform", "win32")
    @patch("cmdbox.runtime.shell.which")
    def test_build_shell_command_windows_preferred_shell_not_found(self, mock_which):
        # preferred_shell is powershell, but it is not found. Fallback to cmd.exe
        mock_which.side_effect = lambda x: x == "cmd.exe"
        command = "echo hello"
        result = build_shell_command(command, preferred_shell="powershell")
        self.assertEqual(result, ["cmd.exe", "/C", "echo hello"])
