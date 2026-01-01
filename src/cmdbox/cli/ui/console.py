from rich.console import Console

from cmdbox.models import Command
from cmdbox.resolve.types import ResolveResult

console = Console()


def print_error(message: str) -> None:
    console.print(f"Error: {message}", style="bold red")


def print_success(message: str) -> None:
    console.print(message, style="bold green")


def print_command(command: Command) -> None:
    console.print(f"[bold green]Alias: [/bold green]{command.alias}")
    console.print(f"[bold purple3]Template: [/bold purple3]{command.template}")
    if command.description:
        console.print(
            f"[bold cornflower_blue]Description: [/bold cornflower_blue]{command.description}"
        )
    tags = command.tags
    if tags:
        for tag in tags:
            console.print(f"[reverse blue]{tag}[/reverse blue]")


def print_command_list(commands: list[Command]) -> None:
    console.print(f"[bold #325ba8]{len(commands)} commands found:\n[/bold #325ba8]")
    for command in commands:
        print_command(command)
        console.print("")


def print_run_preview(result: ResolveResult) -> None:
    console.print(f"[bold green]Preview:[/bold green] {result.text}")
    for step in result.trace:
        console.print(step.kind)
        console.print(step.key)
        console.print(step.expanded_to)
