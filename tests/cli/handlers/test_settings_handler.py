import unittest
from unittest.mock import MagicMock, patch
import typer
from cmdbox.cli.handlers.settings_handler import run_edit_settings, run_open_data_dir
from cmdbox.settings.settings_service import SettingsService
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.editor import EditCanceled
from cmdbox.core.paths import APP_DATA_DIR


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


class TestRunOpenDataDir(unittest.TestCase):

    def setUp(self):
        self.mock_console = MagicMock(spec=ConsoleUI)
        self.get_console = lambda: self.mock_console

    @patch("cmdbox.cli.handlers.settings_handler.subprocess.run")
    @patch("os.startfile", create=True)
    def test_print_only_prints_path_and_does_not_open(self, mock_startfile, mock_run):
        run_open_data_dir(print_only=True, get_console=self.get_console)

        self.mock_console.print.assert_called_once_with(APP_DATA_DIR)
        mock_run.assert_not_called()
        mock_startfile.assert_not_called()

    @patch("cmdbox.cli.handlers.settings_handler.subprocess.run")
    @patch("os.startfile", create=True)
    @patch("cmdbox.cli.handlers.settings_handler.platform.system")
    def test_windows_uses_os_startfile(self, mock_system, mock_startfile, mock_run):
        mock_system.return_value = "Windows"

        run_open_data_dir(print_only=False, get_console=self.get_console)

        mock_startfile.assert_called_once_with(APP_DATA_DIR)
        mock_run.assert_not_called()

    @patch("cmdbox.cli.handlers.settings_handler.subprocess.run")
    @patch("os.startfile", create=True)
    @patch("cmdbox.cli.handlers.settings_handler.platform.system")
    def test_macos_uses_open_command(self, mock_system, mock_startfile, mock_run):
        mock_system.return_value = "Darwin"

        run_open_data_dir(print_only=False, get_console=self.get_console)

        mock_run.assert_called_once_with(["open", str(APP_DATA_DIR)], check=True)
        mock_startfile.assert_not_called()

    @patch("cmdbox.cli.handlers.settings_handler.subprocess.run")
    @patch("os.startfile", create=True)
    @patch("cmdbox.cli.handlers.settings_handler.platform.system")
    def test_linux_uses_xdg_open(self, mock_system, mock_startfile, mock_run):
        mock_system.return_value = "Linux"

        run_open_data_dir(print_only=False, get_console=self.get_console)

        mock_run.assert_called_once_with(["xdg-open", str(APP_DATA_DIR)], check=True)
        mock_startfile.assert_not_called()

    @patch("cmdbox.cli.handlers.settings_handler.subprocess.run")
    @patch("os.startfile", create=True)
    @patch("cmdbox.cli.handlers.settings_handler.platform.system")
    def test_get_console_not_called_when_not_print_only(
        self, mock_system, mock_startfile, mock_run
    ):
        mock_system.return_value = "Linux"
        console_spy = MagicMock(wraps=self.get_console)

        run_open_data_dir(print_only=False, get_console=console_spy)

        console_spy.assert_not_called()
