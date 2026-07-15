import unittest
from unittest.mock import MagicMock, patch

from rich.text import Text

from cmdbox.cli.ui.console import ConsoleUI


class TestConsole(unittest.TestCase):

    def _console_ui(self, **kwargs):
        defaults = dict(pager_mode="auto", pager_min_rows=25)
        defaults.update(kwargs)
        ui = ConsoleUI(theme=MagicMock(), **defaults)
        return ui

    def test_force_true_always_pages_even_when_not_terminal(self):
        ui = self._console_ui()
        ui._console = MagicMock(is_terminal=False)
        self.assertTrue(ui._should_page(row_count=1, force=True))

    def test_force_false_never_pages_even_in_always_mode(self):
        ui = self._console_ui(pager_mode="always")
        ui._console = MagicMock(is_terminal=True)
        self.assertFalse(ui._should_page(row_count=100, force=False))

    def test_never_pages_when_not_a_terminal(self):
        ui = self._console_ui(pager_mode="always")
        ui._console = MagicMock(is_terminal=False)
        self.assertFalse(ui._should_page(row_count=100, force=None))

    def test_never_mode_does_not_page(self):
        ui = self._console_ui(pager_mode="never")
        ui._console = MagicMock(is_terminal=True)
        self.assertFalse(ui._should_page(row_count=100, force=None))

    def test_always_mode_pages_regardless_of_row_count(self):
        ui = self._console_ui(pager_mode="always")
        ui._console = MagicMock(is_terminal=True)
        self.assertTrue(ui._should_page(row_count=1, force=None))

    def test_auto_mode_pages_above_threshold(self):
        ui = self._console_ui(pager_mode="auto", pager_min_rows=25)
        ui._console = MagicMock(is_terminal=True)
        self.assertTrue(ui._should_page(row_count=26, force=None))

    def test_auto_mode_does_not_page_at_or_below_threshold(self):
        ui = self._console_ui(pager_mode="auto", pager_min_rows=25)
        ui._console = MagicMock(is_terminal=True)
        self.assertFalse(ui._should_page(row_count=25, force=None))

    def test_auto_mode_pages_when_row_count_unknown(self):
        ui = self._console_ui(pager_mode="auto")
        ui._console = MagicMock(is_terminal=True)
        self.assertTrue(ui._should_page(row_count=None, force=None))

    def test_prints_directly_when_should_page_is_false(self):
        ui = self._console_ui()
        ui._should_page = MagicMock(return_value=False)
        ui._console = MagicMock()
        renderable = MagicMock()

        ui.print_paged(renderable, row_count=5)

        ui._console.print.assert_called_once_with(renderable)

    @patch("cmdbox.cli.ui.console.Pager")
    def test_routes_to_pager_when_should_page_is_true(self, mock_pager_cls):
        ui = self._console_ui(pager_page_step=15, pager_line_step=1)
        ui._should_page = MagicMock(return_value=True)
        ui._render_to_ansi = MagicMock(return_value="rendered ansi")

        ui.print_paged(MagicMock(), row_count=100)

        mock_pager_cls.assert_called_once_with(page_step=15, line_step=1)
        mock_pager_cls.return_value.run_pager.assert_called_once_with("rendered ansi")

    def test_render_to_ansi_produces_nonempty_string(self):
        ui = self._console_ui(use_color=True)
        result = ui._render_to_ansi(Text("hello"))
        self.assertIn("hello", result)
