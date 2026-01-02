from typing import Sequence

from rich.console import Console

from cmdbox.models import Command, Variable
from cmdbox.resolve.types import ResolveResult


class ConsoleUI:

    def __init__(self, theme, *, force_color=None):
        self._console = Console(
            theme=theme.rich, force_terminal=force_color, highlight=False
        )
        self._theme = theme

    def print(self, message: str) -> None:
        self._console.print(message)

    def success(self, message: str) -> None:
        self._console.print(message, style=self._theme.success)

    def warning(self, message: str) -> None:
        self._console.print(message, style=self._theme.warning)

    def error(self, message: str) -> None:
        self._console.print(message, style=self._theme.error)

    def info(self, message: str) -> None:
        self._console.print(message, style=self._theme.info)

    def muted(self, message: str) -> None:
        self._console.print(message, style=self._theme.muted)

    def debug(self, message: str) -> None:
        self._console.print(message, style=self._theme.debug)

    def print_command(
        self, command: Command, output_fields: Sequence[str] | None = None
    ) -> None:
        display_map = [
            ("alias", "Alias", command.alias, self._theme.command_alias),
            ("template", "Template", command.template, self._theme.command_template),
            (
                "description",
                "Description",
                command.description,
                self._theme.command_description,
            ),
            (
                "created",
                "Created",
                command.date_created,
                self._theme.command_date_created,
            ),
            (
                "updated",
                "Updated",
                command.last_updated,
                self._theme.command_last_updated,
            ),
            ("used", "Used", command.used, self._theme.command_used),
            (
                "last_used",
                "Last used",
                command.last_used,
                self._theme.command_last_used,
            ),
        ]

        for field_key, label, value, style in display_map:
            if not output_fields or field_key in output_fields:
                self._console.print(f"{label}: {value}", style=style)

    def print_command_list(
        self, commands: list[Command], output_fields: Sequence[str] | None = None
    ) -> None:
        self.success(f"{len(commands)} commands found:\n")
        for command in commands:
            self.print_command(command, output_fields=output_fields)
            self.print("")

    def print_run_preview(self, result: ResolveResult) -> None:
        self._console.print(result.text, style=self._theme.run_preview_command)
        for step in result.trace:
            self._console.print(step.kind, style=self._theme.run_preview_step_kind)
            self._console.print(step.key, style=self._theme.run_preview_step_key)
            self._console.print(
                step.expanded_to, style=self._theme.run_preview_step_expanded_to
            )

    def print_variable(
        self, var: Variable, output_fields: Sequence[str] | None = None
    ) -> None:
        display_map = [
            ("name", "Name", var.name, self._theme.variable_name),
            ("value", "Value", var.value, self._theme.variable_value),
            ("created", "Created", var.date_created, self._theme.variable_date_created),
            ("updated", "Updated", var.last_updated, self._theme.variable_last_updated),
        ]

        for field_key, label, value, style in display_map:
            if not output_fields or field_key in output_fields:
                self._console.print(f"{label}: {value}", style=style)

    def print_variable_list(
        self, variables: list[Variable], output_fields: Sequence[str] | None = None
    ) -> None:
        self.success(f"{len(variables)} variables found:\n")
        for var in variables:
            self.print_variable(var, output_fields=output_fields)
            self.print("")
