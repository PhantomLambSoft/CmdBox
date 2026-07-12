from prompt_toolkit import Application
from prompt_toolkit.formatted_text import ANSI
from prompt_toolkit.key_binding import KeyBindings
from prompt_toolkit.layout import HSplit, Layout, Window
from prompt_toolkit.layout.controls import FormattedTextControl

_HELP_TEXT = " j/k or ↑/↓ scroll · ctrl-d/ctrl-u page · g/G top/bottom · q quit "
_PAGE_STEP = 15


def run_pager(ansi_text: str) -> None:
    """
    Displays ansi_text in a full-screen, scrollable pager. Blocks until the user quits.

    Args:
        ansi_text: Pre-rendered ANSI text (typically captured from a Rich Console)
            to display.
    """
    content_control = FormattedTextControl(ANSI(ansi_text))
    content_window = Window(
        content=content_control, wrap_lines=False, always_hide_cursor=True
    )

    footer_window = Window(
        content=FormattedTextControl(_HELP_TEXT),
        height=1,
        style="reverse",
    )

    kb = KeyBindings()

    @kb.add("q")
    @kb.add("c-c")
    @kb.add("escape")
    def _quit(event) -> None:
        event.app.exit()

    @kb.add("j")
    @kb.add("down")
    def _down(event) -> None:
        content_window.vertical_scroll += 1

    @kb.add("k")
    @kb.add("up")
    def _up(event) -> None:
        content_window.vertical_scroll = max(0, content_window.vertical_scroll - 1)

    @kb.add("c-d")
    @kb.add("pagedown")
    def _page_down(event) -> None:
        content_window.vertical_scroll += _PAGE_STEP

    @kb.add("c-u")
    @kb.add("pageup")
    def _page_up(event) -> None:
        content_window.vertical_scroll = max(
            0, content_window.vertical_scroll - _PAGE_STEP
        )

    @kb.add("g")
    def _top(event) -> None:
        content_window.vertical_scroll = 0

    @kb.add("G")
    def _bottom(event) -> None:
        content_window.vertical_scroll = 10**9  # clamps to max content on render

    layout = Layout(HSplit([content_window, footer_window]))

    Application(
        layout=layout,
        key_bindings=kb,
        full_screen=True,
        mouse_support=True,
    ).run()
