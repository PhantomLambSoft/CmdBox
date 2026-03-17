import unittest
from unittest.mock import patch, MagicMock

import typer

from cmdbox.cli.commands import command_run
from cmdbox.cli.handlers.run_handler import RawRunContext


class TestCommandRun(unittest.TestCase):

    def make_mock_ctx(self, args: list[str] | None = None) -> MagicMock:
        ctx = MagicMock(spec=typer.Context)
        ctx.args = args or []
        return ctx

    @patch("cmdbox.cli.commands.command_run.run_run_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_run_default(self, mock_container, mock_run_run_command):
        # Setup
        alias = "test-alias"

        # Execute
        command_run.run(alias=alias, ctx=self.make_mock_ctx())

        # Verify
        mock_run_run_command.assert_called_once()
        args, kwargs = mock_run_run_command.call_args
        self.assertEqual(kwargs["alias"], alias)
        self.assertIsInstance(kwargs["run_ctx"], RawRunContext)
        self.assertIsNone(kwargs["run_ctx"].cwd)
        self.assertIsNone(kwargs["run_ctx"].env)
        self.assertFalse(kwargs["run_ctx"].capture)
        self.assertIsNone(kwargs["run_ctx"].shell)
        self.assertIsNone(kwargs["run_ctx"].verbose)
        self.assertEqual(kwargs["get_run_service"], mock_container.get_run_service)
        self.assertEqual(kwargs["get_console"], mock_container.get_console)
        self.assertEqual(kwargs["get_settings"], mock_container.get_settings)

    @patch("cmdbox.cli.commands.command_run.run_run_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_run_with_options(self, mock_container, mock_run_run_command):
        # Setup
        alias = "test-alias"
        runtime_vars = {"VAR1": "VAL1", "VAR2": "VAL2"}
        cwd = "/tmp"
        env = ["VAR1=VAL1", "VAR2=VAL2"]
        capture = True
        shell = "bash"
        verbose = True

        # Execute
        command_run.run(
            alias=alias,
            cwd=cwd,
            env=env,
            capture=capture,
            shell=shell,
            verbose=verbose,
            ctx=self.make_mock_ctx(),
        )

        # Verify
        mock_run_run_command.assert_called_once()
        kwargs = mock_run_run_command.call_args[1]
        self.assertEqual(kwargs["alias"], alias)
        self.assertEqual(kwargs["run_ctx"].cwd, cwd)
        self.assertEqual(kwargs["run_ctx"].env, env)
        self.assertTrue(kwargs["run_ctx"].capture)
        self.assertEqual(kwargs["run_ctx"].shell, shell)
        self.assertTrue(kwargs["run_ctx"].verbose)

    @patch("cmdbox.cli.commands.command_run.preview")
    def test_run_with_preview_flag(self, mock_preview):
        # Setup
        alias = "test-alias"
        cwd = "/tmp"
        env = ["VAR1=VAL1"]
        capture = True
        shell = "zsh"
        verbose = True

        mock_ctx = self.make_mock_ctx()

        # Execute
        command_run.run(
            alias=alias,
            preview_cmd=True,
            cwd=cwd,
            env=env,
            capture=capture,
            shell=shell,
            verbose=verbose,
            ctx=mock_ctx,
        )

        # Verify
        mock_preview.assert_called_once_with(
            alias=alias,
            cwd=cwd,
            env=env,
            capture=capture,
            shell=shell,
            verbose=verbose,
            ctx=mock_ctx,
        )

    @patch("cmdbox.cli.commands.command_run.run_preview_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_preview_default(self, mock_container, mock_run_preview_command):
        # Setup
        alias = "test-alias"

        # Execute
        command_run.preview(alias=alias, ctx=self.make_mock_ctx())

        # Verify
        mock_run_preview_command.assert_called_once()
        kwargs = mock_run_preview_command.call_args[1]
        self.assertEqual(kwargs["alias"], alias)
        self.assertIsInstance(kwargs["run_ctx"], RawRunContext)
        self.assertIsNone(kwargs["run_ctx"].cwd)
        self.assertIsNone(kwargs["run_ctx"].env)
        self.assertFalse(kwargs["run_ctx"].capture)
        self.assertIsNone(kwargs["run_ctx"].shell)
        self.assertIsNone(kwargs["run_ctx"].verbose)
        self.assertEqual(kwargs["get_run_service"], mock_container.get_run_service)
        self.assertEqual(kwargs["get_console"], mock_container.get_console)
        self.assertEqual(kwargs["get_settings"], mock_container.get_settings)

    @patch("cmdbox.cli.commands.command_run.run_preview_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_preview_with_options(self, mock_container, mock_run_preview_command):
        # Setup
        alias = "test-alias"
        cwd = "/tmp"
        env = ["VAR1=VAL1"]
        capture = True
        shell = "fish"
        verbose = True

        # Execute
        command_run.preview(
            alias=alias,
            cwd=cwd,
            env=env,
            capture=capture,
            shell=shell,
            verbose=verbose,
            ctx=self.make_mock_ctx(),
        )

        # Verify
        mock_run_preview_command.assert_called_once()
        kwargs = mock_run_preview_command.call_args[1]
        self.assertEqual(kwargs["alias"], alias)
        self.assertEqual(kwargs["run_ctx"].cwd, cwd)
        self.assertEqual(kwargs["run_ctx"].env, env)
        self.assertTrue(kwargs["run_ctx"].capture)
        self.assertEqual(kwargs["run_ctx"].shell, shell)
        self.assertTrue(kwargs["run_ctx"].verbose)

    @patch("cmdbox.cli.commands.command_run.run_run_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_run_empty_env(self, mock_container, mock_run_run_command):
        # Setup
        alias = "test-alias"
        env = []

        # Execute
        command_run.run(alias=alias, env=env, ctx=self.make_mock_ctx())

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
        command_run.preview(
            alias=alias,
            cwd=None,
            env=None,
            capture=False,
            shell=None,
            ctx=self.make_mock_ctx(),
        )

        # Verify
        mock_run_preview_command.assert_called_once()
        kwargs = mock_run_preview_command.call_args[1]
        self.assertIsNone(kwargs["run_ctx"].cwd)
        self.assertIsNone(kwargs["run_ctx"].env)
        self.assertFalse(kwargs["run_ctx"].capture)
        self.assertIsNone(kwargs["run_ctx"].shell)

    @patch("cmdbox.cli.commands.command_run.run_run_command")
    @patch("cmdbox.cli.commands.command_run.container")
    def test_runtime_var_parsing(self, mock_container, mock_run_run_command):
        alias = "test-alias"

        command_run.run(alias="test-alias", ctx=self.make_mock_ctx(["--name", "Homer"]))

        mock_run_run_command.assert_called_once()
        args, kwargs = mock_run_run_command.call_args
        self.assertEqual(kwargs["alias"], alias)
        self.assertEqual(kwargs["runtime_vars"], {"name": "Homer"})
