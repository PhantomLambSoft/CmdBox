import unittest
from unittest.mock import patch
import os
from cmdbox.cli.ui.editor import (
    resolve_editor,
    edit_text_in_editor,
    edit_text_fullscreen,
    EditCanceled,
)


class TestEditor(unittest.TestCase):

    @patch.dict(os.environ, {"VISUAL": "code --wait", "EDITOR": "nano"}, clear=True)
    def test_resolve_editor_visual(self):
        self.assertEqual(["code", "--wait"], resolve_editor())

    @patch.dict(os.environ, {"EDITOR": "vim"}, clear=True)
    def test_resolve_editor_editor(self):
        self.assertEqual(["vim"], resolve_editor())

    @patch.dict(os.environ, {}, clear=True)
    @patch("os.name", "nt")
    def test_resolve_editor_windows_default(self):
        self.assertEqual(["notepad"], resolve_editor())

    @patch.dict(os.environ, {}, clear=True)
    @patch("os.name", "posix")
    @patch("shutil.which")
    def test_resolve_editor_posix_candidates(self, mock_which):
        def side_effect(cmd):
            return "/usr/bin/" + cmd if cmd == "nano" else None

        mock_which.side_effect = side_effect
        self.assertEqual(["nano"], resolve_editor())

    @patch.dict(os.environ, {}, clear=True)
    @patch("os.name", "posix")
    @patch("shutil.which", return_value=None)
    def test_resolve_editor_posix_fallback(self, mock_which):
        self.assertEqual(["vi"], resolve_editor())

    @patch("cmdbox.cli.ui.editor.resolve_editor", return_value=["myeditor"])
    @patch("subprocess.run")
    @patch("tempfile.TemporaryDirectory")
    @patch("pathlib.Path.write_text")
    @patch("pathlib.Path.read_text", return_value="edited text")
    def test_edit_text_in_editor_success(
        self, mock_read, mock_write, mock_tempdir, mock_run, mock_resolve
    ):
        mock_tempdir.return_value.__enter__.return_value = "temp_dir"

        result = edit_text_in_editor("initial text", suffix=".toml", title_hint="Hint")

        self.assertEqual("edited text", result)
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        self.assertEqual(["myeditor", os.path.join("temp_dir", "edit.toml")], args[0])
        self.assertEqual("Hint", kwargs["env"]["CB_EDIT_TITLE"])

    @patch("cmdbox.cli.ui.editor.resolve_editor", return_value=["nonexistent"])
    @patch("subprocess.run")
    def test_edit_text_in_editor_not_found(self, mock_run, mock_resolve):
        mock_run.side_effect = FileNotFoundError()
        with self.assertRaises(RuntimeError):
            edit_text_in_editor("text")

    @patch("cmdbox.cli.ui.editor.Application")
    def test_edit_text_fullscreen_success(self, mock_app_class):
        mock_app = mock_app_class.return_value
        mock_app.run.return_value = "edited content"

        result = edit_text_fullscreen("initial", title="My Title")

        self.assertEqual("edited content", result)
        mock_app.run.assert_called_once()

    @patch("cmdbox.cli.ui.editor.Application")
    def test_edit_text_fullscreen_cancel(self, mock_app_class):
        mock_app = mock_app_class.return_value
        mock_app.run.side_effect = EditCanceled()

        with self.assertRaises(EditCanceled):
            edit_text_fullscreen("initial")
