import click
from typer.core import TyperGroup


class AliasFallbackGroup(TyperGroup):
    """
    A class that allows a "default" command to be run when the app is run with no
    existing command name. This lets the run command be the default used command,
    so users can run a command by only specifying its alias, without having to
    use the "run" command.
    """

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
        rv = super().get_command(ctx, cmd_name)
        if rv is not None:
            return rv

        run_cmd = super().get_command(ctx, "run")
        if run_cmd is None:
            return None

        forwarded_params = [p for p in run_cmd.params if p.name != "alias"]

        @click.command(
            cmd_name, params=forwarded_params, help=f"Run stored command '{cmd_name}'."
        )
        @click.pass_context
        def _alias_cmd(inner_ctx: click.Context, **kwargs):
            inner_ctx.invoke(run_cmd, alias=cmd_name, **kwargs)

        return _alias_cmd
