from rich.console import Console


class ConsoleUI:

    def __init__(self, theme, *, force_color=None):
        self._console = Console(
            theme=theme, force_terminal=force_color, highlight=False
        )
        self._theme = theme

    def print(self, thing) -> None:
        self._console.print(thing)

    def success(self, message: str) -> None:
        self._console.print(message, style="status.success")

    def warning(self, message: str) -> None:
        self._console.print(message, style=self._theme.warning)

    def error(self, message: str) -> None:
        self._console.print(message, style="status.error")

    def info(self, message: str) -> None:
        self._console.print(message, style=self._theme.info)

    def muted(self, message: str) -> None:
        self._console.print(message, style=self._theme.muted)

    def debug(self, message: str) -> None:
        self._console.print(message, style=self._theme.debug)
