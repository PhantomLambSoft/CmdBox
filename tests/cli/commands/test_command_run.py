import unittest
from unittest.mock import patch
from cmdbox.cli.commands import command_run
from cmdbox.cli.handlers.run_handler import RawRunContext


class TestCommandRun(unittest.TestCase):

    @patch("cmdbox.cli.commands.command_run.run_run_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_run_default(self, mock_container, mock_run_run_command):
        # Setup
        alias = "test-alias"

        # Execute
        command_run.run(alias=alias)

        # Verify
        mock_run_run_command.assert_called_once()
        args, kwargs = mock_run_run_command.call_args
        self.assertEqual(kwargs["alias"], alias)
        self.assertIsInstance(kwargs["run_ctx"], RawRunContext)
        self.assertIsNone(kwargs["run_ctx"].cwd)
        self.assertIsNone(kwargs["run_ctx"].env)
        self.assertFalse(kwargs["run_ctx"].capture)
        self.assertIsNone(kwargs["run_ctx"].shell)
        self.assertEqual(kwargs["get_run_service"], mock_container.get_run_service)
        self.assertEqual(kwargs["get_console"], mock_container.get_console)

    @patch("cmdbox.cli.commands.command_run.run_run_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_run_with_options(self, mock_container, mock_run_run_command):
        # Setup
        alias = "test-alias"
        cwd = "/tmp"
        env = ["VAR1=VAL1", "VAR2=VAL2"]
        capture = True
        shell = "bash"

        # Execute
        command_run.run(alias=alias, cwd=cwd, env=env, capture=capture, shell=shell)

        # Verify
        mock_run_run_command.assert_called_once()
        kwargs = mock_run_run_command.call_args[1]
        self.assertEqual(kwargs["alias"], alias)
        self.assertEqual(kwargs["run_ctx"].cwd, cwd)
        self.assertEqual(kwargs["run_ctx"].env, env)
        self.assertTrue(kwargs["run_ctx"].capture)
        self.assertEqual(kwargs["run_ctx"].shell, shell)

    @patch("cmdbox.cli.commands.command_run.preview")
    def test_run_with_preview_flag(self, mock_preview):
        # Setup
        alias = "test-alias"
        cwd = "/tmp"
        env = ["VAR1=VAL1"]
        capture = True
        shell = "zsh"

        # Execute
        command_run.run(
            alias=alias,
            preview_cmd=True,
            cwd=cwd,
            env=env,
            capture=capture,
            shell=shell,
        )

        # Verify
        mock_preview.assert_called_once_with(
            alias=alias, cwd=cwd, env=env, capture=capture, shell=shell
        )

    @patch("cmdbox.cli.commands.command_run.run_preview_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_preview_default(self, mock_container, mock_run_preview_command):
        # Setup
        alias = "test-alias"

        # Execute
        command_run.preview(alias=alias)

        # Verify
        mock_run_preview_command.assert_called_once()
        kwargs = mock_run_preview_command.call_args[1]
        self.assertEqual(kwargs["alias"], alias)
        self.assertIsInstance(kwargs["run_ctx"], RawRunContext)
        self.assertIsNone(kwargs["run_ctx"].cwd)
        self.assertIsNone(kwargs["run_ctx"].env)
        self.assertFalse(kwargs["run_ctx"].capture)
        self.assertIsNone(kwargs["run_ctx"].shell)
        self.assertEqual(kwargs["get_run_service"], mock_container.get_run_service)
        self.assertEqual(kwargs["get_console"], mock_container.get_console)

    @patch("cmdbox.cli.commands.command_run.run_preview_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_preview_with_options(self, mock_container, mock_run_preview_command):
        # Setup
        alias = "test-alias"
        cwd = "/tmp"
        env = ["VAR1=VAL1"]
        capture = True
        shell = "fish"

        # Execute
        command_run.preview(alias=alias, cwd=cwd, env=env, capture=capture, shell=shell)

        # Verify
        mock_run_preview_command.assert_called_once()
        kwargs = mock_run_preview_command.call_args[1]
        self.assertEqual(kwargs["alias"], alias)
        self.assertEqual(kwargs["run_ctx"].cwd, cwd)
        self.assertEqual(kwargs["run_ctx"].env, env)
        self.assertTrue(kwargs["run_ctx"].capture)
        self.assertEqual(kwargs["run_ctx"].shell, shell)

    @patch("cmdbox.cli.commands.command_run.run_run_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_run_empty_env(self, mock_container, mock_run_run_command):
        # Setup
        alias = "test-alias"
        env = []

        # Execute
        command_run.run(alias=alias, env=env)

        # Verify
        mock_run_run_command.assert_called_once()
        kwargs = mock_run_run_command.call_args[1]
        self.assertEqual(kwargs["run_ctx"].env, [])

    @patch("cmdbox.cli.commands.command_run.run_preview_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_preview_none_values(self, mock_container, mock_run_preview_command):
        # Setup
        alias = "test-alias"

        # Execute
        command_run.preview(alias=alias, cwd=None, env=None, capture=False, shell=None)

        # Verify
        mock_run_preview_command.assert_called_once()
        kwargs = mock_run_preview_command.call_args[1]
        self.assertIsNone(kwargs["run_ctx"].cwd)
        self.assertIsNone(kwargs["run_ctx"].env)
        self.assertFalse(kwargs["run_ctx"].capture)
        self.assertIsNone(kwargs["run_ctx"].shell)
