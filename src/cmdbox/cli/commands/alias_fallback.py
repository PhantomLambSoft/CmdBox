import click
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
    }

    def get_command(self, ctx: click.Context, cmd_name: str):
        """
        Retrieves a command based on the command name, creating an alias command if
        the command was not directly found.

        If the command corresponding to `cmd_name` doesn't exist, a new alias
        command is generated using the `run` command, forwarding all its parameters
        except the `alias` parameter.

        Args:
            ctx (click.Context): The Click context in which the command is being
                invoked.
            cmd_name (str): The name of the command to retrieve.

        Returns:
            click.Command: The command object corresponding to the given
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

        forwarded_params = [p for p in run_cmd.params if p.name != "alias"]

        @click.command(
            cmd_name,
            params=forwarded_params,
            help=f"Run stored command '{cmd_name}'.",
            context_settings=run_cmd.context_settings,
        )
        @click.pass_context
        def _alias_cmd(inner_ctx: click.Context, **kwargs):
            inner_ctx.meta["_extra_args"] = inner_ctx.args[:]
            inner_ctx.invoke(run_cmd, alias=cmd_name, **kwargs)

        return _alias_cmd
