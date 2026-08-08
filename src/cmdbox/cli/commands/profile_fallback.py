import typer
from typer.core import TyperGroup


class ProfileFallback(TyperGroup):
    """
    Represents a custom command group with fallback behavior for resolving commands.

    This class extends the functionality of TyperGroup by providing a fallback
    mechanism that redirects unrecognized commands to a specified command. If the
    first argument in the list of arguments matches a valid command, it proceeds as
    usual. Otherwise, the fallback mechanism appends the "switch" command and
    attempts to resolve the resulting command sequence.

    Attributes:
        name (str): The name of the command group.
        help (Optional[str]): An optional help message for the command group.
    """

    def resolve_command(self, ctx: typer.Context, args: list):
        """
        Resolves a command to handle user input and determine the appropriate
        action to execute.

        The method first checks whether no arguments are passed or if the given
        command name exists in the available commands. If so, it resolves the
        command normally using the parent class's implementation. Otherwise, it
        prepends "switch" to the arguments list and resolves the command with
        the modified input. This approach allows for dynamic command handling
        and seamless fallback mechanisms.

        Args:
            ctx (typer.Context): The context object associated with the command,
                containing metadata and execution context for the command-line
                utility.
            args (list): A list of command-line arguments passed by the user.
                The first element typically represents the command name, with
                subsequent elements comprising the command's arguments.

        Returns:
            Optional[Command]: The resolved command object if a valid command is
                determined based on the given arguments; otherwise, it defers to
                the resolution mechanism of the superclass.
        """
        if not args:
            return super().resolve_command(ctx, args)

        cmd_name = args[0]

        if super().get_command(ctx, cmd_name) is not None:
            return super().resolve_command(ctx, args)

        return super().resolve_command(ctx, ["switch"] + list(args))
