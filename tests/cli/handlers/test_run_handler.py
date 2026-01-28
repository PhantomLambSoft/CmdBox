import unittest
from unittest.mock import MagicMock, patch
import typer
from cmdbox.cli.handlers.run_handler import (
    RawRunContext,
    run_run_command,
    run_preview_command,
    get_run_ctx,
    parse_env,
)
from cmdbox.runtime.executor import RunContext
from cmdbox.runtime.results import ExecutionResult
from cmdbox.resolve.types import ResolveResult


class TestRunHandler(unittest.TestCase):

    def test_parse_env_none(self):
        self.assertIsNone(parse_env(None))

    def test_parse_env_single_str(self):
        self.assertEqual(parse_env("KEY=VALUE"), {"KEY": "VALUE"})

    def test_parse_env_comma_separated_str(self):
        self.assertEqual(
            parse_env("KEY1=VALUE1,KEY2=VALUE2"), {"KEY1": "VALUE1", "KEY2": "VALUE2"}
        )

    def test_parse_env_list(self):
        self.assertEqual(
            parse_env(["KEY1=VALUE1", "KEY2=VALUE2"]),
            {"KEY1": "VALUE1", "KEY2": "VALUE2"},
        )

    def test_parse_env_mixed_list_and_comma(self):
        self.assertEqual(
            parse_env(["KEY1=VALUE1,KEY2=VALUE2", "KEY3=VALUE3"]),
            {"KEY1": "VALUE1", "KEY2": "VALUE2", "KEY3": "VALUE3"},
        )

    def test_parse_env_invalid_format(self):
        with self.assertRaises(typer.BadParameter) as cm:
            parse_env("INVALID_FORMAT")
        self.assertIn("Invalid environment variable format", str(cm.exception))

    def test_parse_env_invalid_format_in_list(self):
        with self.assertRaises(typer.BadParameter) as cm:
            parse_env(["KEY1=VALUE1", "INVALID_FORMAT"])
        self.assertIn("Invalid environment variable format", str(cm.exception))

    def test_get_run_ctx_none(self):
        ctx = get_run_ctx(None)
        self.assertIsInstance(ctx, RunContext)
        self.assertIsNone(ctx.cwd)
        self.assertIsNone(ctx.env)
        self.assertFalse(ctx.capture)
        self.assertIsNone(ctx.shell)

    def test_get_run_ctx_valid(self):
        raw_ctx = RawRunContext(cwd="/tmp", env="KEY=VALUE", capture=True, shell="bash")
        ctx = get_run_ctx(raw_ctx)
        self.assertEqual(ctx.cwd, "/tmp")
        self.assertEqual(ctx.env, {"KEY": "VALUE"})
        self.assertTrue(ctx.capture)
        self.assertEqual(ctx.shell, "bash")

    @patch("cmdbox.cli.handlers.run_handler.render_execution_result")
    def test_run_run_command(self, mock_render):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello", stderr=""
        )
        mock_run_service.run.return_value = mock_ex_result
        mock_render.return_value = "rendered_result"

        run_run_command(
            alias="test-alias",
            run_ctx=None,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.run.assert_called_once_with("test-alias", ctx=RunContext())
        mock_render.assert_called_once_with(mock_ex_result)
        mock_console.print.assert_called_once_with("rendered_result")

    @patch("cmdbox.cli.handlers.run_handler.render_execution_result")
    def test_run_run_command_with_error(self, mock_render):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=1, stdout="", stderr="error occurred"
        )
        mock_run_service.run.return_value = mock_ex_result
        mock_render.return_value = "rendered_error"

        run_run_command(
            alias="test-alias",
            run_ctx=None,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_render.assert_called_once_with(mock_ex_result)
        mock_console.print.assert_called_once_with("rendered_error")

    def test_run_run_command_with_ctx(self):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello", stderr=""
        )
        mock_run_service.run.return_value = mock_ex_result
        raw_ctx = RawRunContext(cwd="/tmp")

        run_run_command(
            alias="test-alias",
            run_ctx=raw_ctx,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.run.assert_called_once_with(
            "test-alias", ctx=RunContext(cwd="/tmp")
        )

    @patch("cmdbox.cli.handlers.run_handler.render_preview_result")
    def test_run_preview_command(self, mock_render):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_resolve_result = ResolveResult(text="echo hello", trace=[])
        mock_run_service.preview.return_value = mock_resolve_result
        mock_render.return_value = "rendered_preview"

        run_preview_command(
            alias="test-alias",
            run_ctx=None,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.preview.assert_called_once_with("test-alias")
        mock_render.assert_called_once_with(mock_resolve_result)
        mock_console.print.assert_called_once_with("rendered_preview")

    @patch("cmdbox.cli.handlers.run_handler.render_preview_result")
    def test_run_preview_command_with_ctx(self, mock_render):
        # Even if run_ctx is passed to run_preview_command, it's currently only used
        # to initialize run_ctx variable, but run_service.preview only takes alias.
        # Let's verify this behavior.
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_resolve_result = ResolveResult(text="echo hello", trace=[])
        mock_run_service.preview.return_value = mock_resolve_result
        raw_ctx = RawRunContext(cwd="/tmp")
        mock_render.return_value = "rendered_preview"

        run_preview_command(
            alias="test-alias",
            run_ctx=raw_ctx,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.preview.assert_called_once_with("test-alias")
        mock_render.assert_called_once_with(mock_resolve_result)
        mock_console.print.assert_called_once_with("rendered_preview")
