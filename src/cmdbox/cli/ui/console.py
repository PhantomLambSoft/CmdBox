import io

from rich.console import Console

from cmdbox.cli.ui.pager import Pager


class ConsoleUI:

    def __init__(
        self,
        theme,
        *,
        use_color: bool = True,
        pager_mode: str = "auto",
        pager_min_rows: int = 25,
    ):
        self._console = Console(theme=theme, no_color=not use_color, highlight=False)
        self._theme = theme
        self._user_color = use_color
        self._pager_mode = pager_mode
        self._pager_min_rows = pager_min_rows

    def print(self, thing) -> None:
        self._console.print(thing)

    def print_paged(
        self, renderable, *, row_count: int | None = None, force: bool | None = None
    ) -> None:
        """
        Prints the given renderable content to the console or a pager based on the conditions
        evaluated by `_should_page`. If paging is not required, the content is directly
        printed. Otherwise, the content is rendered to ANSI text and displayed in a pager.

        Args:
            renderable: The content to be rendered. Can be any object compatible with
                the console's print functionality.
            row_count: Optional; The number of rows to use as a condition for determining
                whether to use the pager. If None, no row-based condition is applied.
            force: Optional; If True, forces the use of a pager regardless of row count
                or other conditions. If None or False, the decision to use the pager
                is made based on other criteria.
        """
        if not self._should_page(row_count=row_count, force=force):
            self._console.print(renderable)
            return

        ansi_text = self._render_to_ansi(renderable)
        pager = Pager()
        pager.run_pager(ansi_text)

    def _should_page(self, *, row_count: int | None, force: bool | None) -> bool:
        if force is not None:
            return force

        if not self._console.is_terminal:
            return False  # Never page a redirected/piped output

        if self._pager_mode == "never":
            return False
        if self._pager_mode == "always":
            return True

        if row_count is None:
            return True
        return row_count > self._pager_min_rows

    def _render_to_ansi(self, renderable) -> str:
        buf = io.StringIO()
        capture = Console(
            file=buf,
            theme=self._theme,
            no_color=not self._user_color,
            highlight=False,
            force_terminal=True,
            width=self._console.width,
        )
        capture.print(renderable)
        return buf.getvalue()

    def success(self, message: str) -> None:
        self._console.print(message, style="status.success")

    def warning(self, message: str) -> None:
        self._console.print(message, style="status.warning")

    def error(self, message: str) -> None:
        self._console.print(message, style="status.error")

    def info(self, message: str) -> None:
        self._console.print(message, style="status.info")

    def muted(self, message: str) -> None:
        self._console.print(message, style="status.muted")

    def debug(self, message: str) -> None:
        self._console.print(message, style="status.debug")
