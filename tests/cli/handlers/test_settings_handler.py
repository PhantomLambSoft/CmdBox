import unittest
from unittest.mock import MagicMock, patch
import typer
from cmdbox.cli.handlers.settings_handler import run_edit_settings
from cmdbox.settings.settings_service import SettingsService
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.editor import EditCanceled


class TestSettingsHandler(unittest.TestCase):

    def setUp(self):
        self.mock_settings_service = MagicMock(spec=SettingsService)
        self.mock_console = MagicMock(spec=ConsoleUI)
        self.get_settings_service = lambda: self.mock_settings_service
        self.get_console = lambda: self.mock_console

    @patch("cmdbox.cli.handlers.settings_handler.edit_text_in_editor")
    def test_run_edit_settings_external_success(self, mock_edit_in_editor):
        mock_edit_in_editor.return_value = "new settings"

        run_edit_settings(
            external=True,
            get_settings_service=self.get_settings_service,
            get_console=self.get_console,
        )

        # Verify that settings_service.edit was called
        self.mock_settings_service.edit.assert_called_once()
        # Verify the editor callback passed to edit() uses edit_text_in_editor
        editor_fn = self.mock_settings_service.edit.call_args[0][0]
        result = editor_fn("initial")
        self.assertEqual("new settings", result)
        mock_edit_in_editor.assert_called_with(
            "initial", suffix=".toml", title_hint="CmdBox Settings"
        )

        self.mock_console.success.assert_called_with("Settings saved.")

    @patch("cmdbox.cli.handlers.settings_handler.edit_text_fullscreen")
    def test_run_edit_settings_internal_success(self, mock_edit_fullscreen):
        mock_edit_fullscreen.return_value = "new settings"

        run_edit_settings(
            external=False,
            get_settings_service=self.get_settings_service,
            get_console=self.get_console,
        )

        self.mock_settings_service.edit.assert_called_once()
        editor_fn = self.mock_settings_service.edit.call_args[0][0]
        result = editor_fn("initial")
        self.assertEqual("new settings", result)
        mock_edit_fullscreen.assert_called_with("initial", title="Edit Settings")

        self.mock_console.success.assert_called_with("Settings saved.")

    def test_run_edit_settings_canceled(self):
        self.mock_settings_service.edit.side_effect = EditCanceled()

        with self.assertRaises(typer.Exit) as cm:
            run_edit_settings(
                get_settings_service=self.get_settings_service,
                get_console=self.get_console,
            )

        self.assertEqual(0, cm.exception.exit_code)
        self.mock_console.success.assert_not_called()

    def test_run_edit_settings_error(self):
        self.mock_settings_service.edit.side_effect = Exception("Boom")

        with self.assertRaises(typer.Exit) as cm:
            run_edit_settings(
                get_settings_service=self.get_settings_service,
                get_console=self.get_console,
            )

        self.assertEqual(1, cm.exception.exit_code)
        self.mock_console.error.assert_called_once()
        self.assertIn("Boom", self.mock_console.error.call_args[0][0])
