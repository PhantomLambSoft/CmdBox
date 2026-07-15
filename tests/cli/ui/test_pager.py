import unittest
from unittest.mock import MagicMock, patch
from cmdbox.cli.ui.pager import Pager


class TestPager(unittest.TestCase):

    def setUp(self):
        self.pager = Pager(page_step=5, line_step=1)
        self.ansi_text = "\n".join([f"line {i}" for i in range(20)])
        self.pager.lines = self.ansi_text.splitlines()
        self.pager.total_lines = len(self.pager.lines)

    @patch("shutil.get_terminal_size")
    def test_visible_height(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)
        # 10 - 1 (footer) = 9
        self.assertEqual(self.pager._visible_height(), 9)

        mock_get_terminal_size.return_value = MagicMock(lines=1)
        # max(1, 1 - 1) = 1
        self.assertEqual(self.pager._visible_height(), 1)

    @patch("shutil.get_terminal_size")
    def test_max_offset(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)
        # total_lines = 20, visible_height = 9. max_offset = 20 - 9 = 11
        self.assertEqual(self.pager._max_offset(), 11)

        mock_get_terminal_size.return_value = MagicMock(lines=30)
        # total_lines = 20, visible_height = 29. max_offset = max(0, 20 - 29) = 0
        self.assertEqual(self.pager._max_offset(), 0)

    @patch("shutil.get_terminal_size")
    def test_clamp(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)  # max_offset = 11
        self.assertEqual(self.pager._clamp(-5), 0)
        self.assertEqual(self.pager._clamp(5), 5)
        self.assertEqual(self.pager._clamp(15), 11)

    @patch("shutil.get_terminal_size")
    def test_down(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)
        event = MagicMock()
        self.pager._down(event)
        self.assertEqual(self.pager.offset, 1)
        event.app.invalidate.assert_called_once()

    @patch("shutil.get_terminal_size")
    def test_up(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)
        self.pager.offset = 5
        event = MagicMock()
        self.pager._up(event)
        self.assertEqual(self.pager.offset, 4)
        event.app.invalidate.assert_called_once()

    @patch("shutil.get_terminal_size")
    def test_page_down(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)
        event = MagicMock()
        self.pager._page_down(event)
        self.assertEqual(self.pager.offset, 5)  # page_step is 5
        event.app.invalidate.assert_called_once()

    @patch("shutil.get_terminal_size")
    def test_page_up(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)
        self.pager.offset = 10
        event = MagicMock()
        self.pager._page_up(event)
        self.assertEqual(self.pager.offset, 5)  # page_step is 5
        event.app.invalidate.assert_called_once()

    @patch("shutil.get_terminal_size")
    def test_top(self, mock_get_terminal_size):
        self.pager.offset = 10
        event = MagicMock()
        self.pager._top(event)
        self.assertEqual(self.pager.offset, 0)
        event.app.invalidate.assert_called_once()

    @patch("shutil.get_terminal_size")
    def test_bottom(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=10)  # max_offset = 11
        event = MagicMock()
        self.pager._bottom(event)
        self.assertEqual(self.pager.offset, 11)
        event.app.invalidate.assert_called_once()

    def test_quit(self):
        event = MagicMock()
        self.pager._quit(event)
        event.app.exit.assert_called_once()

    @patch("shutil.get_terminal_size")
    def test_get_visible_text(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=5)  # visible_height = 4
        self.pager.offset = 2
        visible_text = self.pager._get_visible_text()
        # lines[2:6] -> line 2, line 3, line 4, line 5
        expected_text = "line 2\nline 3\nline 4\nline 5"
        self.assertEqual(visible_text.value, expected_text)

    @patch("shutil.get_terminal_size")
    def test_status_text(self, mock_get_terminal_size):
        mock_get_terminal_size.return_value = MagicMock(lines=5)  # visible_height = 4
        self.pager.offset = 2
        status = self.pager._status_text()
        # top = 2+1=3, bottom = min(2+4, 20) = 6
        self.assertIn("line 3-6 of 20", status)

    @patch("cmdbox.cli.ui.pager.Application")
    @patch("cmdbox.cli.ui.pager.Layout")
    @patch("cmdbox.cli.ui.pager.HSplit")
    @patch("cmdbox.cli.ui.pager.Window")
    @patch("cmdbox.cli.ui.pager.FormattedTextControl")
    def test_run_pager(self, mock_ftc, mock_window, mock_hsplit, mock_layout, mock_app):
        # We need HSplit to return something that to_container accepts if it's not mocked,
        # but since we are mocking Layout and it receives the HSplit result,
        # it should be fine as long as HSplit is also mocked.
        self.pager.run_pager(self.ansi_text)
        self.assertEqual(self.pager.total_lines, 20)
        mock_app.assert_called_once()
        mock_app.return_value.run.assert_called_once()

    def test_key_bindings(self):
        kb = self.pager.kb
        self.assertTrue(kb.get_bindings_for_keys(("q",)))
        self.assertTrue(kb.get_bindings_for_keys(("c-c",)))
        self.assertTrue(kb.get_bindings_for_keys(("escape",)))
        self.assertTrue(kb.get_bindings_for_keys(("j",)))
        self.assertTrue(kb.get_bindings_for_keys(("down",)))
        self.assertTrue(kb.get_bindings_for_keys(("k",)))
        self.assertTrue(kb.get_bindings_for_keys(("up",)))
        self.assertTrue(kb.get_bindings_for_keys(("c-d",)))
        self.assertTrue(kb.get_bindings_for_keys(("pagedown",)))
        self.assertTrue(kb.get_bindings_for_keys(("c-u",)))
        self.assertTrue(kb.get_bindings_for_keys(("pageup",)))
        self.assertTrue(kb.get_bindings_for_keys(("g",)))
        self.assertTrue(kb.get_bindings_for_keys(("G",)))
