import unittest
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

    # --- Direct command resolution ---

    def test_direct_command_resolution_uses_existing_command(self):
        app, calls = self._build_app()

        result = self.runner.invoke(app, ["status"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("status_called", False))
        self.assertNotIn("alias", calls)

    # --- _command_aliases mapping ---

    def test_alias_mapping_resolves_short_name_to_real_command(self):
        app, calls = self._build_app()

        result = self.runner.invoke(app, ["tags"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("tag_called", False))
        self.assertNotIn("alias", calls)

    def test_all_configured_aliases_resolve_to_their_target_command(self):
        for alias, target in AliasFallbackGroup._command_aliases.items():
            with self.subTest(alias=alias, target=target):
                calls = {}
                app = typer.Typer(cls=AliasFallbackGroup)

                @app.command(target)
                def _target_cmd():
                    calls["called"] = True

                @app.command("noop")
                def _noop_cmd():
                    pass

                result = self.runner.invoke(app, [alias])

                self.assertEqual(0, result.exit_code)
                self.assertTrue(calls.get("called", False))

    def test_alias_group_class_is_still_used_with_only_one_other_command(self):
        """
        Regression test: Typer collapses a Typer() app down to a plain
        single-command CLI when only one command is registered, which skips
        building a Group entirely and means `cls=AliasFallbackGroup` never
        gets instantiated. AliasFallbackGroup only takes effect once there
        are 2+ commands, forcing Typer to build a real Group.
        """
        calls = {}
        app = typer.Typer(cls=AliasFallbackGroup)

        @app.command("cmd")
        def _cmd():
            calls["cmd_called"] = True

        @app.command("status")
        def _status():
            pass

        result = self.runner.invoke(app, ["cmds"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("cmd_called", False))

    # --- Shortcut command expansion (resolve_command) ---

    def test_shortcut_command_expands_to_full_subcommand(self):
        calls = {}
        app = typer.Typer(cls=AliasFallbackGroup)
        history_app = typer.Typer()

        @history_app.command("last")
        def last():
            calls["last_called"] = True

        app.add_typer(history_app, name="history")

        result = self.runner.invoke(app, ["!!"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("last_called", False))

    # --- Dynamically generated alias (fallback) commands ---

    def test_fallback_command_is_generated_from_run_command_metadata(self):
        app, _ = self._build_app(with_tag=False)
        typer_group = get_command(app)
        ctx = typer.Context(typer_group)

        run_cmd = typer_group.get_command(ctx, "run")
        generated_cmd = typer_group.get_command(ctx, "deploy")

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

    def test_generated_alias_command_preserves_default_option_value(self):
        """
        Regression test: _build_alias_command rebuilds the function signature
        via inspect.Signature to strip the `alias` parameter. This confirms
        that rebuild doesn't lose the original default value for `mode`, or
        introduce spurious extra_args when none were passed.
        """
        app, calls = self._build_app(with_tag=False)

        result = self.runner.invoke(app, ["deploy", "artifact"])

        self.assertEqual(0, result.exit_code)
        self.assertEqual("fast", calls["mode"])
        self.assertEqual([], calls["extra_args"])

    def test_fallback_still_works_alongside_alias_mapping(self):
        app, calls = self._build_app()  # both run and tag present

        result = self.runner.invoke(app, ["deploy", "artifact"])

        self.assertEqual(0, result.exit_code)
        self.assertEqual("deploy", calls["alias"])

    def test_get_command_returns_none_when_run_command_is_missing(self):
        app, _ = self._build_app(with_run=False, with_tag=True)
        typer_group = get_command(app)
        ctx = typer.Context(typer_group)

        cmd = typer_group.get_command(ctx, "deploy")

        self.assertIsNone(cmd)
