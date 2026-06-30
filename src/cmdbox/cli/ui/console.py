from rich.console import Console


class ConsoleUI:

    def __init__(self, theme, *, use_color: bool = True):
        self._console = Console(theme=theme, no_color=not use_color, highlight=False)
        self._theme = theme

    def print(self, thing) -> None:
        self._console.print(thing)

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
