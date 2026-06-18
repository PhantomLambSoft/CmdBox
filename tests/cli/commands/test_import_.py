import unittest
from pathlib import Path
from unittest.mock import patch

from cmdbox.cli.commands import import_


class TestImportCommand(unittest.TestCase):

    @patch("cmdbox.cli.commands.import_.run_import_file")
    @patch("cmdbox.cli.commands.import_.container")
    def test_import_file_defaults(self, mock_container, mock_run_import_file):
        path = Path("import.json")

        import_.import_file(path=path)

        mock_run_import_file.assert_called_once_with(
            path=path,
            overwrite=False,
            preview=False,
            get_import_service=mock_container.get_import_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.import_.run_import_file")
    @patch("cmdbox.cli.commands.import_.container")
    def test_import_file_forwards_true_flags(
        self, mock_container, mock_run_import_file
    ):
        path = Path("configs\\import.json")

        import_.import_file(path=path, overwrite=True, preview=True)

        mock_run_import_file.assert_called_once_with(
            path=path,
            overwrite=True,
            preview=True,
            get_import_service=mock_container.get_import_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.import_.run_import_file")
    @patch("cmdbox.cli.commands.import_.container")
    def test_import_file_preserves_path_object(
        self, mock_container, mock_run_import_file
    ):
        path = Path("C:\\tmp\\exports\\data.json")

        import_.import_file(path=path, overwrite=False, preview=True)

        kwargs = mock_run_import_file.call_args.kwargs
        self.assertEqual(path, kwargs["path"])
        self.assertEqual(False, kwargs["overwrite"])
        self.assertEqual(True, kwargs["preview"])
        self.assertEqual(
            mock_container.get_import_service, kwargs["get_import_service"]
        )
        self.assertEqual(mock_container.get_console, kwargs["get_console"])
