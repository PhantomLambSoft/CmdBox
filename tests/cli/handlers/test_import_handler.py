import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

import typer

from cmdbox.cli.handlers.import_handler import run_import_file
from cmdbox.services.errors import ImportCycleError, ImportFileError
from cmdbox.services.import_service import ImportResult


class TestImportHandler(unittest.TestCase):

    def setUp(self):
        self.mock_import_service = MagicMock()
        self.mock_console = MagicMock()
        self.get_import_service = lambda: self.mock_import_service
        self.get_console = lambda: self.mock_console

    @patch("cmdbox.cli.handlers.import_handler.render_import_result")
    def test_run_import_file_prints_import_result_when_preview_is_false(
        self, mock_render
    ):
        path = Path("import.json")
        result = ImportResult(commands_created=["build"], variables_created=["ENV"])
        self.mock_import_service.import_file.return_value = result
        mock_render.return_value = "rendered_result"

        run_import_file(
            path=path,
            overwrite=False,
            preview=False,
            profile=None,
            get_import_service=self.get_import_service,
            get_console=self.get_console,
        )

        self.mock_import_service.import_file.assert_called_once_with(
            path,
            overwrite=False,
            preview=False,
            profile=None,
        )
        mock_render.assert_called_once_with(result)
        self.mock_console.print.assert_called_once_with("rendered_result")
        self.mock_console.error.assert_not_called()

    @patch("cmdbox.cli.handlers.import_handler.render_import_preview")
    @patch("cmdbox.cli.handlers.import_handler.render_import_result")
    def test_run_import_file_prints_preview_when_preview_is_true(
        self, mock_render_result, mock_render_preview
    ):
        path = Path("configs\\import.json")
        result = ImportResult(commands_skipped=["build"], variables_overwritten=["ENV"])
        self.mock_import_service.import_file.return_value = result
        mock_render_preview.return_value = "rendered_preview"

        run_import_file(
            path=path,
            overwrite=True,
            preview=True,
            profile=None,
            get_import_service=self.get_import_service,
            get_console=self.get_console,
        )

        self.mock_import_service.import_file.assert_called_once_with(
            path,
            overwrite=True,
            preview=True,
            profile=None,
        )
        mock_render_preview.assert_called_once_with(result, source=str(path))
        mock_render_result.assert_not_called()
        self.mock_console.print.assert_called_once_with("rendered_preview")
        self.mock_console.error.assert_not_called()

    def test_run_import_file_raises_exit_for_import_file_error(self):
        path = Path("bad.json")
        self.mock_import_service.import_file.side_effect = ImportFileError(
            "Could not read import file"
        )

        with self.assertRaises(typer.Exit) as context:
            run_import_file(
                path=path,
                overwrite=False,
                preview=False,
                profile=None,
                get_import_service=self.get_import_service,
                get_console=self.get_console,
            )

        self.assertEqual(1, context.exception.exit_code)
        self.mock_console.error.assert_called_once_with("Could not read import file")
        self.mock_console.print.assert_not_called()

    def test_run_import_file_raises_exit_for_import_cycle_error(self):
        path = Path("cyclic.json")
        cycle = ["cmd:build", "var:ENV", "cmd:build"]
        self.mock_import_service.import_file.side_effect = ImportCycleError(cycle)

        with self.assertRaises(typer.Exit) as context:
            run_import_file(
                path=path,
                overwrite=True,
                preview=True,
                profile=None,
                get_import_service=self.get_import_service,
                get_console=self.get_console,
            )

        self.assertEqual(1, context.exception.exit_code)
        self.mock_console.error.assert_called_once_with(
            "Import rejected - circular dependency detected: "
            "cmd:build -> var:ENV -> cmd:build"
        )
        self.mock_console.print.assert_not_called()

    def test_run_import_file_with_profile(self):
        path = Path("import.json")
        profile = "test-profile"
        self.mock_import_service.import_file.return_value = ImportResult()

        run_import_file(
            path=path,
            overwrite=False,
            preview=False,
            profile=profile,
            get_import_service=self.get_import_service,
            get_console=self.get_console,
        )

        self.mock_import_service.import_file.assert_called_once_with(
            path,
            overwrite=False,
            preview=False,
            profile=profile,
        )
