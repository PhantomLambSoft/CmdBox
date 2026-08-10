import unittest
from unittest.mock import MagicMock, patch, call
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
from cmdbox.resolve.type_defs import ResolveResult


class TestRunHandler(unittest.TestCase):

    def setUp(self):
        self.settings = MagicMock()
        self.settings.execution_settings.default_verbose = False
        self.settings.execution_settings.capture_output = False
        self.settings.execution_settings.default_shell = None

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
        self.assertIsNone(ctx.capture)
        self.assertIsNone(ctx.shell)
        self.assertIsNone(ctx.verbose)

    def test_get_run_ctx_valid(self):
        raw_ctx = RawRunContext(
            cwd="/tmp", env="KEY=VALUE", capture=True, shell="bash", verbose=True
        )
        ctx = get_run_ctx(raw_ctx)
        self.assertEqual(ctx.cwd, "/tmp")
        self.assertEqual(ctx.env, {"KEY": "VALUE"})
        self.assertTrue(ctx.capture)
        self.assertEqual(ctx.shell, "bash")
        self.assertTrue(ctx.verbose)

    @patch("cmdbox.cli.handlers.run_handler.render_execution_result")
    def test_run_run_command_with_verbose(self, mock_render):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello", stderr=""
        )
        mock_run_service.run.return_value = mock_ex_result
        mock_run_service.collect_missing_vars.return_value = []
        mock_render.return_value = "rendered_result"

        run_run_command(
            alias="test-alias",
            runtime_vars=None,
            run_ctx=RawRunContext(verbose=True),
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=None
        )
        mock_run_service.run.assert_called_once_with(
            "test-alias", runtime_vars=None, ctx=RunContext(verbose=True), profile=None
        )
        mock_render.assert_called_once_with(mock_ex_result)
        mock_console.print.assert_called_once_with("rendered_result")

    @patch("cmdbox.cli.handlers.run_handler.render_execution_result")
    def test_run_run_command_no_verbose(self, mock_render):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello", stderr=""
        )
        mock_run_service.run.return_value = mock_ex_result
        mock_run_service.collect_missing_vars.return_value = []

        run_run_command(
            alias="test-alias",
            runtime_vars=None,
            run_ctx=RawRunContext(verbose=False),
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=None
        )
        mock_run_service.run.assert_called_once_with(
            "test-alias", ctx=RunContext(verbose=False), runtime_vars=None, profile=None
        )
        mock_render.assert_not_called()
        mock_console.print.assert_not_called()

    @patch("cmdbox.cli.handlers.run_handler.render_execution_result")
    def test_run_run_command_with_error(self, mock_render):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=1, stdout="", stderr="error occurred"
        )
        mock_run_service.run.return_value = mock_ex_result
        mock_run_service.collect_missing_vars.return_value = []
        mock_render.return_value = "rendered_error"

        run_run_command(
            alias="test-alias",
            run_ctx=RawRunContext(verbose=True),
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=None
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
        mock_run_service.collect_missing_vars.return_value = []
        raw_ctx = RawRunContext(cwd="/tmp")

        run_run_command(
            alias="test-alias",
            runtime_vars=None,
            run_ctx=raw_ctx,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=None
        )
        mock_run_service.run.assert_called_once_with(
            "test-alias", runtime_vars=None, ctx=RunContext(cwd="/tmp"), profile=None
        )

    def test_run_run_command_with_runtime_vars(self):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello", stderr=""
        )
        mock_run_service.run.return_value = mock_ex_result
        mock_run_service.collect_missing_vars.return_value = []
        raw_ctx = RawRunContext(cwd="/tmp")

        vars = {"VAR1": "value1", "VAR2": "value2"}

        run_run_command(
            alias="test-alias",
            runtime_vars=vars,
            run_ctx=raw_ctx,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=vars, profile=None
        )
        mock_run_service.run.assert_called_once_with(
            "test-alias", runtime_vars=vars, ctx=RunContext(cwd="/tmp"), profile=None
        )

    @patch("cmdbox.cli.handlers.run_handler.prompt_for_missing_var")
    def test_run_run_command_prompts_for_missing_vars(self, mock_prompt):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello", stderr=""
        )
        mock_run_service.collect_missing_vars.return_value = ["name", "title"]
        mock_run_service.run.return_value = mock_ex_result
        mock_prompt.side_effect = ["Homer", "Colonel"]
        runtime_vars = {}

        run_run_command(
            alias="test-alias",
            runtime_vars=runtime_vars,
            run_ctx=RawRunContext(verbose=False),
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        self.assertEqual({"name": "Homer", "title": "Colonel"}, runtime_vars)
        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=runtime_vars, profile=None
        )
        mock_prompt.assert_has_calls([call("name"), call("title")])
        mock_run_service.run.assert_called_once_with(
            "test-alias",
            ctx=RunContext(verbose=False),
            runtime_vars=runtime_vars,
            profile=None,
        )

    def test_run_run_command_with_profile(self):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_ex_result = ExecutionResult(
            command="echo hello", exit_code=0, stdout="hello", stderr=""
        )
        mock_run_service.run.return_value = mock_ex_result
        mock_run_service.collect_missing_vars.return_value = []
        profile = "test-profile"

        run_run_command(
            alias="test-alias",
            profile=profile,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=profile
        )
        mock_run_service.run.assert_called_once_with(
            "test-alias", runtime_vars=None, ctx=RunContext(), profile=profile
        )

    @patch("cmdbox.cli.handlers.run_handler.get_run_ctx")
    @patch("cmdbox.cli.handlers.run_handler.render_preview_result")
    def test_run_preview_command(self, mock_render, mock_get_run_ctx):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_resolve_result = ResolveResult(text="echo hello", trace=[])
        mock_ctx = MagicMock()
        mock_get_run_ctx.return_value = mock_ctx
        mock_run_service.preview.return_value = mock_resolve_result, mock_ctx
        mock_run_service.collect_missing_vars.return_value = []
        mock_render.return_value = "rendered_preview"

        run_preview_command(
            alias="test-alias",
            run_ctx=mock_ctx,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
            runtime_vars=None,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=None
        )
        mock_run_service.preview.assert_called_once_with(
            "test-alias", runtime_vars=None, ctx=mock_ctx, profile=None
        )
        mock_render.assert_called_once_with(mock_resolve_result, ctx=mock_ctx)
        mock_console.print.assert_called_once_with("rendered_preview")

    @patch("cmdbox.cli.handlers.run_handler.get_run_ctx")
    @patch("cmdbox.cli.handlers.run_handler.render_preview_result")
    def test_run_preview_command_with_profile(self, mock_render, mock_get_run_ctx):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_resolve_result = ResolveResult(text="echo hello", trace=[])
        mock_ctx = MagicMock()
        mock_get_run_ctx.return_value = mock_ctx
        mock_run_service.preview.return_value = mock_resolve_result, mock_ctx
        mock_run_service.collect_missing_vars.return_value = []
        profile = "test-profile"
        mock_render.return_value = "rendered_preview"

        run_preview_command(
            alias="test-alias",
            profile=profile,
            run_ctx=mock_ctx,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=profile
        )
        mock_run_service.preview.assert_called_once_with(
            "test-alias", runtime_vars=None, ctx=mock_ctx, profile=profile
        )
        mock_render.assert_called_once_with(mock_resolve_result, ctx=mock_ctx)
        mock_console.print.assert_called_once_with("rendered_preview")

    @patch("cmdbox.cli.handlers.run_handler.get_run_ctx")
    @patch("cmdbox.cli.handlers.run_handler.render_preview_result")
    def test_run_preview_command_with_ctx(self, mock_render, mock_get_run_ctx):
        self.settings.execution_settings.verbose = True
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_resolve_result = ResolveResult(text="echo hello", trace=[])
        mock_ctx = MagicMock()
        mock_get_run_ctx.return_value = mock_ctx
        mock_run_service.preview.return_value = mock_resolve_result, mock_ctx
        mock_run_service.collect_missing_vars.return_value = []
        raw_ctx = RawRunContext(cwd="/tmp")
        mock_render.return_value = "rendered_preview"

        run_preview_command(
            alias="test-alias",
            run_ctx=raw_ctx,
            runtime_vars=None,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=None, profile=None
        )
        mock_run_service.preview.assert_called_once_with(
            "test-alias", runtime_vars=None, ctx=mock_ctx, profile=None
        )
        mock_get_run_ctx.assert_called_once_with(raw_ctx)
        mock_render.assert_called_once_with(mock_resolve_result, ctx=mock_ctx)
        mock_console.print.assert_called_once_with("rendered_preview")

    @patch("cmdbox.cli.handlers.run_handler.get_run_ctx")
    @patch("cmdbox.cli.handlers.run_handler.prompt_for_missing_var")
    @patch("cmdbox.cli.handlers.run_handler.render_preview_result")
    def test_run_preview_command_prompts_for_missing_vars(
        self, mock_render, mock_prompt, mock_get_run_ctx
    ):
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_resolve_result = ResolveResult(text="echo hello", trace=[])
        mock_run_service.collect_missing_vars.return_value = ["name"]
        mock_ctx = MagicMock()
        mock_get_run_ctx.return_value = mock_ctx
        mock_run_service.preview.return_value = mock_resolve_result, mock_ctx
        mock_prompt.return_value = "Homer"
        mock_render.return_value = "rendered_preview"
        runtime_vars = {}

        run_preview_command(
            alias="test-alias",
            run_ctx=mock_ctx,
            runtime_vars=runtime_vars,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        self.assertEqual({"name": "Homer"}, runtime_vars)
        mock_run_service.collect_missing_vars.assert_called_once_with(
            "test-alias", runtime_vars=runtime_vars, profile=None
        )
        mock_prompt.assert_called_once_with("name")
        mock_run_service.preview.assert_called_once_with(
            "test-alias", runtime_vars=runtime_vars, ctx=mock_ctx, profile=None
        )
        mock_render.assert_called_once_with(mock_resolve_result, ctx=mock_ctx)
        mock_console.print.assert_called_once_with("rendered_preview")
