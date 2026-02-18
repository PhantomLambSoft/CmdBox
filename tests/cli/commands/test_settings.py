import unittest
from unittest.mock import patch
from typer.testing import CliRunner
from cmdbox.cli.commands.settings import app


class TestSettingsCommand(unittest.TestCase):
    def setUp(self):
        self.runner = CliRunner()

    @patch("cmdbox.cli.commands.settings.container.get_settings_service")
    @patch("cmdbox.cli.commands.settings.container.get_console")
    @patch("cmdbox.cli.commands.settings.settings_handler.run_edit_settings")
    def test_settings_edit_default(
        self, mock_run_edit, mock_get_console, mock_get_settings
    ):
        result = self.runner.invoke(app, [])

        self.assertEqual(0, result.exit_code)
        mock_run_edit.assert_called_once()
        args, kwargs = mock_run_edit.call_args
        self.assertFalse(args[0])  # external=False by default in command
        self.assertEqual(mock_get_settings, kwargs["get_settings_service"])
        self.assertEqual(mock_get_console, kwargs["get_console"])

    @patch("cmdbox.cli.commands.settings.container.get_settings_service")
    @patch("cmdbox.cli.commands.settings.container.get_console")
    @patch("cmdbox.cli.commands.settings.settings_handler.run_edit_settings")
    def test_settings_edit_external(
        self, mock_run_edit, mock_get_console, mock_get_settings
    ):
        result = self.runner.invoke(app, ["--external"])

        self.assertEqual(0, result.exit_code)
        mock_run_edit.assert_called_once()
        args, kwargs = mock_run_edit.call_args
        self.assertTrue(args[0])  # external=True

    @patch("cmdbox.cli.commands.settings.container.get_settings_service")
    @patch("cmdbox.cli.commands.settings.container.get_console")
    @patch("cmdbox.cli.commands.settings.settings_handler.run_edit_settings")
    def test_settings_edit_external_short(
        self, mock_run_edit, mock_get_console, mock_get_settings
    ):
        result = self.runner.invoke(app, ["-e"])

        self.assertEqual(0, result.exit_code)
        mock_run_edit.assert_called_once()
        args, kwargs = mock_run_edit.call_args
        self.assertTrue(args[0])
