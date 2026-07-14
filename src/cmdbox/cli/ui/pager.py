import shutil

from prompt_toolkit import Application
from prompt_toolkit.formatted_text import ANSI
from prompt_toolkit.key_binding import KeyBindings
from prompt_toolkit.layout import HSplit, Layout, Window
from prompt_toolkit.layout.controls import FormattedTextControl

_HELP_TEXT = " j/k or ↑/↓ scroll · ctrl-d/ctrl-u page · g/G top/bottom · q quit "


class Pager:
    """
    A full-screen, scrollable pager for pre-rendered ANSI text. Instantiate once per
    call to `run_pager()`. Call `run_pager()` to block until the user quits.

    Scrolling implementation note:
    The obvious approach here is to let Window manage scrolling via its vertical_scroll
    attribute, incrementing or decrement it directly in each key handler. This does not work:
    Window recalculates vertical_scroll on every render to keep its cursor in view, and
    FormattedTextControl has no cursor. With no cursor to track, Window treats the cursor
    position as the top of the content and resets vertical_scroll to 0 on every redraw,
    ignoring the value set by the key handler.

    Instead, scrolling is tracked manually as a plain offset in the pre-split list
    of lines. Each key handler adjusts the offset, and Window is only ever shown the
    slice of lines currently visible. So there is no Window-managed scroll state for
    it to reset.
    """

    def __init__(self, page_step: int = 15, line_step: int = 1) -> None:
        self.page_step = page_step
        self.line_step = line_step
        self.lines: list[str] = []
        self.total_lines = 0
        self.offset = 0
        self.kb = self._build_key_bindings()

    def _build_key_bindings(self) -> KeyBindings:
        kb = KeyBindings()

        kb.add("q")(self._quit)
        kb.add("c-c")(self._quit)
        kb.add("escape")(self._quit)

        kb.add("j")(self._down)
        kb.add("down")(self._down)

        kb.add("k")(self._up)
        kb.add("up")(self._up)

        kb.add("c-d")(self._page_down)
        kb.add("pagedown")(self._page_down)

        kb.add("c-u")(self._page_up)
        kb.add("pageup")(self._page_up)

        kb.add("g")(self._top)
        kb.add("G")(self._bottom)

        return kb

    def _quit(self, event) -> None:
        event.app.exit()

    def _down(self, event) -> None:
        self.offset = self._clamp(self.offset + self.line_step)
        event.app.invalidate()

    def _up(self, event) -> None:
        self.offset = self._clamp(self.offset - self.line_step)
        event.app.invalidate()

    def _page_down(self, event) -> None:
        self.offset = self._clamp(self.offset + self.page_step)
        event.app.invalidate()

    def _page_up(self, event) -> None:
        self.offset = self._clamp(self.offset - self.page_step)
        event.app.invalidate()

    def _top(self, event) -> None:
        self.offset = 0
        event.app.invalidate()

    def _bottom(self, event) -> None:
        self.offset = self._max_offset()
        event.app.invalidate()

    def _visible_height(self) -> int:
        return max(1, shutil.get_terminal_size().lines - 1)  # -1 for footer

    def _max_offset(self) -> int:
        return max(0, self.total_lines - self._visible_height())

    def _clamp(self, offset: int) -> int:
        return max(0, min(offset, self._max_offset()))

    def _get_visible_text(self):
        height = self._visible_height()
        return ANSI("\n".join(self.lines[self.offset : self.offset + height]))

    def _status_text(self):
        top = self.offset + 1
        bottom = min(self.offset + self._visible_height(), self.total_lines)
        return f" line {top}-{bottom} of {self.total_lines} · {_HELP_TEXT}"

    def run_pager(self, ansi_text: str) -> None:
        """
        Processes ANSI text, splits it into individual lines, and initializes a
        full-screen application with a content display window and a footer window.
        The method creates a scrollable layout and displays the processed text
        with additional functionality such as key bindings and mouse support.

        Args:
            ansi_text (str): The ANSI-formatted string to be displayed in the
                pager application. The text is split into lines for rendering.
        """
        self.lines = ansi_text.splitlines()
        self.total_lines = len(self.lines)

        content_window = Window(
            content=FormattedTextControl(self._get_visible_text),
            wrap_lines=False,
            always_hide_cursor=True,
        )
        footer_window = Window(
            content=FormattedTextControl(self._status_text),
            height=1,
            style="reverse",
        )

        layout = Layout(HSplit([content_window, footer_window]))

        Application(
            layout=layout,
            key_bindings=self.kb,
            full_screen=True,
            mouse_support=True,
        ).run()
