import unittest
import click
import typer
from typer.main import get_command
from typer.testing import CliRunner

from cmdbox.cli.commands.alias_fallback import AliasFallbackGroup


class TestAliasFallbackGroup(unittest.TestCase):

    def setUp(self):
        self.runner = CliRunner()

    def _build_app(self, with_run: bool = True, with_tag: bool = True):
        calls = {}
        app = typer.Typer(cls=AliasFallbackGroup)

        if with_run:

            @app.command(
                context_settings={
                    "allow_extra_args": True,
                    "ignore_unknown_options": True,
                }
            )
            def run(
                ctx: typer.Context,
                alias: str,
                item: str,
                mode: str = typer.Option("fast", "--mode"),
            ):
                calls["alias"] = alias
                calls["item"] = item
                calls["mode"] = mode
                calls["extra_args"] = list(ctx.meta.get("_extra_args", []))

        if with_tag:

            @app.command("tag")
            def tag():
                calls["tag_called"] = True

        @app.command("status")
        def status():
            calls["status_called"] = True

        return app, calls

    def test_direct_command_resolution_uses_existing_command(self):
        app, calls = self._build_app()

        result = self.runner.invoke(app, ["status"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("status_called", False))
        self.assertNotIn("alias", calls)

    def test_alias_does_not_forward_to_run(self):
        app, calls = self._build_app()

        self.runner.invoke(app, ["tags"])

        self.assertNotIn("alias", calls)
        self.assertEqual("tags", calls.get("alias", "tags"))

    def test_fallback_command_is_generated_from_run_command_metadata(self):
        app, _ = self._build_app(with_tag=False)
        click_app = get_command(app)
        ctx = click.Context(click_app)

        run_cmd = click_app.get_command(ctx, "run")
        generated_cmd = click_app.get_command(ctx, "deploy")

        self.assertIsNotNone(run_cmd)
        self.assertIsNotNone(generated_cmd)
        self.assertEqual("deploy", generated_cmd.name)
        self.assertEqual("Run stored command 'deploy'.", generated_cmd.help)
        self.assertEqual(run_cmd.context_settings, generated_cmd.context_settings)

        generated_param_names = [param.name for param in generated_cmd.params]
        self.assertNotIn("alias", generated_param_names)
        self.assertIn("item", generated_param_names)
        self.assertIn("mode", generated_param_names)

    def test_generated_alias_command_forwards_args_and_injects_alias(self):
        app, calls = self._build_app(with_tag=False)

        result = self.runner.invoke(
            app,
            ["deploy", "artifact", "--mode", "safe", "--unknown", "42"],
        )

        self.assertEqual(0, result.exit_code)
        self.assertEqual("deploy", calls["alias"])
        self.assertEqual("artifact", calls["item"])
        self.assertEqual("safe", calls["mode"])
        self.assertEqual(["--unknown", "42"], calls["extra_args"])

    def test_get_command_returns_none_when_run_command_is_missing(self):
        app, _ = self._build_app(with_run=False, with_tag=True)
        click_app = get_command(app)
        ctx = click.Context(click_app)

        cmd = click_app.get_command(ctx, "deploy")

        self.assertIsNone(cmd)

    def test_fallback_still_works_alongside_alias_mapping(self):
        app, calls = self._build_app()  # both run and tag present

        result = self.runner.invoke(app, ["deploy", "artifact"])

        self.assertEqual(0, result.exit_code)
        self.assertEqual("deploy", calls["alias"])
