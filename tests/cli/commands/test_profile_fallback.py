import unittest
import typer
from typer.testing import CliRunner
from cmdbox.cli.commands.profile_fallback import ProfileFallback


class TestProfileFallback(unittest.TestCase):

    def setUp(self):
        self.runner = CliRunner()

    def _build_app(self, no_args_is_help=False):
        calls = {}
        app = typer.Typer(cls=ProfileFallback, no_args_is_help=no_args_is_help)
        return app, calls

    def test_resolve_normal_command(self):
        app, calls = self._build_app()

        @app.command("status")
        def status():
            calls["status_called"] = True

        @app.command("switch")
        def switch(
            name: str,
            cmd: bool = typer.Option(False, "--cmd"),
            var: bool = typer.Option(False, "--var"),
        ):
            calls["switch_called"] = True
            calls["name"] = name

        result = self.runner.invoke(app, ["status"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("status_called", False))
        self.assertNotIn("switch_called", calls)

    def test_fallback_to_switch(self):
        app, calls = self._build_app()

        @app.command("switch")
        def switch(name: str):
            calls["switch_called"] = True
            calls["name"] = name

        result = self.runner.invoke(app, ["my-profile"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("switch_called", False))
        self.assertEqual("my-profile", calls.get("name"))

    def test_fallback_to_switch_with_options(self):
        app, calls = self._build_app()

        @app.command("switch")
        def switch(
            name: str,
            cmd: bool = typer.Option(False, "--cmd"),
            var: bool = typer.Option(False, "--var"),
        ):
            calls["switch_called"] = True
            calls["name"] = name
            calls["cmd"] = cmd
            calls["var"] = var

        result = self.runner.invoke(app, ["my-profile", "--cmd", "--var"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("switch_called", False))
        self.assertEqual("my-profile", calls.get("name"))
        self.assertTrue(calls.get("cmd"))
        self.assertTrue(calls.get("var"))

    def test_no_args_resolution(self):
        # When no args, it should trigger Typer's default behavior.
        app, calls = self._build_app()

        @app.command("status")
        def status():
            calls["status_called"] = True

        # In Typer 0.12+, a group with no default command returns 2 when no args are passed.
        # But here we want to make sure it doesn't crash or trigger the fallback.
        result = self.runner.invoke(app, [])

        self.assertNotIn("switch_called", calls)

    def test_fallback_with_multiple_args(self):
        app, calls = self._build_app()

        @app.command("multi")
        def multi(arg1: str, arg2: str):
            calls["multi_called"] = True
            calls["arg1"] = arg1
            calls["arg2"] = arg2

        # "multi" is not a command, so it falls back to "switch"
        # but "switch" only takes 1 argument.
        # Let's test fallback to a command that DOES exist but is prepended.
        # Actually the fallback ALWAYS prepends "switch".

        # If I pass "my-profile", it becomes "switch" "my-profile".
        # If I pass "my-profile" "something", it becomes "switch" "my-profile" "something".

        @app.command("switch")
        def switch_multi(name: str, extra: str = typer.Argument(None)):
            calls["switch_called"] = True
            calls["name"] = name
            calls["extra"] = extra

        result = self.runner.invoke(app, ["my-profile", "extra-arg"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("switch_called", False))
        self.assertEqual("my-profile", calls.get("name"))
        self.assertEqual("extra-arg", calls.get("extra"))

    def test_fallback_does_not_double_switch(self):
        app, calls = self._build_app()

        @app.command("switch")
        def switch(name: str):
            calls["switch_called"] = True
            calls["name"] = name

        # Define status so the app isn't a single-command CLI (which behaves differently)
        @app.command("status")
        def status():
            pass

        result = self.runner.invoke(app, ["switch", "my-profile"])

        self.assertEqual(0, result.exit_code)
        self.assertTrue(calls.get("switch_called", False))
        self.assertEqual("my-profile", calls.get("name"))
