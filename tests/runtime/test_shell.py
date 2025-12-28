import unittest
from unittest.mock import patch
import os
import sys
from cmdbox.runtime.shell import build_shell_command


class TestShell(unittest.TestCase):

    @patch("sys.platform", "win")
    @patch.dict(os.environ, {"COMSPEC": "cmd.exe"}, clear=True)
    def test_build_shell_command_windows_default(self):
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["cmd.exe", "/c", "echo hello"])

    @patch("sys.platform", "win")
    @patch.dict(os.environ, {"CMDBOX_SHELL": "powershell"}, clear=True)
    def test_build_shell_command_windows_powershell(self):
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["pwsh", "-NoProfile", "-Command", "echo hello"])

    @patch("sys.platform", "win")
    @patch.dict(os.environ, {"CMDBOX_SHELL": "pwsh"}, clear=True)
    def test_build_shell_command_windows_pwsh(self):
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["pwsh", "-NoProfile", "-Command", "echo hello"])

    @patch("sys.platform", "linux")
    @patch.dict(os.environ, {"SHELL": "/bin/bash"}, clear=True)
    def test_build_shell_command_linux_bash(self):
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["/bin/bash", "-lc", "echo hello"])

    @patch("sys.platform", "linux")
    @patch.dict(os.environ, {}, clear=True)
    def test_build_shell_command_linux_default(self):
        command = "echo hello"
        result = build_shell_command(command)
        self.assertEqual(result, ["/bin/sh", "-lc", "echo hello"])

    @patch("sys.platform", "win")
    @patch.dict(os.environ, {"COMSPEC": "C:\\Windows\\system32\\cmd.exe"}, clear=True)
    def test_build_shell_command_windows_custom_comspec(self):
        command = "dir"
        result = build_shell_command(command)
        self.assertEqual(result, ["C:\\Windows\\system32\\cmd.exe", "/c", "dir"])
