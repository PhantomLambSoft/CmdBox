import inspect
import typer
from typer.core import TyperGroup


class AliasFallbackGroup(TyperGroup):
    """
    Handles command grouping with support for alias-based fallback.

    This class extends the functionality of TyperGroup to include command alias
    resolution. When a command is not directly found, it attempts to resolve the
    command name using a predefined alias mapping and dynamically generates an
    alias command that forwards its arguments to a core `run` command. This allows
    users to define shorter or alternate names for commands while maintaining
    flexibility.

    Attributes:
        _command_aliases (dict[str, str]): A mapping of alias names to their
            corresponding actual command names. Used for resolving commands
            through aliases.
    """

    _command_aliases: dict[str, str] = {
        "cmds": "cmd",
        "vars": "var",
        "tags": "tag",
        "hist": "history",
    }

    """
    The second item in the list is the name of the command as configured by
    `@app.command("command_name")`, not the method name. Further items in the list
    will be supplied to that sub-command as args.
    """
    _shortcut_commands: dict[str, list[str]] = {"!!": ["history", "last"]}

    def get_command(self, ctx: typer.Context, cmd_name: str):
        """
        Retrieves a command based on the command name, creating an alias command if
        the command was not directly found.

        If the command corresponding to `cmd_name` doesn't exist, a new alias
        command is generated using the `run` command, forwarding all its parameters
        except the `alias` parameter.

        Args:
            ctx (typer.Context): The typer context in which the command is being
                invoked.
            cmd_name (str): The name of the command to retrieve.

        Returns:
            typer.Command: The command object corresponding to the given
            `cmd_name`, or an alias command if the original command does not exist.
            Returns None if `run` command is also unavailable.
        """
        cmd_name = self._command_aliases.get(cmd_name, cmd_name)

        rv = super().get_command(ctx, cmd_name)
        if rv is not None:
            return rv

        run_cmd = super().get_command(ctx, "run")
        if run_cmd is None:
            return None

        return self._build_alias_command(cmd_name, run_cmd)

    def _build_alias_command(self, cmd_name: str, run_cmd):
        original_func = run_cmd.callback
        sig = inspect.signature(original_func)

        visible_params = [
            p for name, p in sig.parameters.items() if name not in ("alias", "ctx")
        ]

        def wrapper(ctx: typer.Context, *args, **kwargs):
            ctx.meta["_extra_args"] = ctx.args[:]
            return original_func(*args, alias=cmd_name, **kwargs)

        wrapper.__name__ = cmd_name
        wrapper.__signature__ = sig.replace(
            parameters=[
                inspect.Parameter(
                    "ctx",
                    inspect.Parameter.POSITIONAL_OR_KEYWORD,
                    annotation=typer.Context,
                ),
            ]
            + visible_params
        )
        wrapper.__doc__ = f"Run stored command '{cmd_name}'."

        temp_app = typer.Typer(add_completion=False)
        temp_app.command(cmd_name, context_settings=run_cmd.context_settings)(wrapper)

        return typer.main.get_command(temp_app)

    def resolve_command(self, ctx: typer.Context, args: list):
        """
        Resolves the command by expanding shortcut commands if applicable.

        This method checks whether the provided command arguments include a shortcut
        command. If a shortcut command is detected, it is expanded into its full form
        and the resulting extended argument list is passed to the parent class's
        `resolve_command` method. If no shortcut is found, the original argument list
        is passed directly.

        Args:
            ctx (typer.Context): The typer context containing information about the
                execution of the command.
            args (list): A list of command-line arguments passed to the command.

        Returns:
            tuple: A tuple containing the command name, the command object, and a list
            of remaining arguments.
        """
        if args and args[0] in self._shortcut_commands:
            expanded = self._shortcut_commands[args[0]]
            return super().resolve_command(ctx, expanded + list(args[1:]))
        return super().resolve_command(ctx, args)
