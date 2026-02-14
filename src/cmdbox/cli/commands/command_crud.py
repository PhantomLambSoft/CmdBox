from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.completions.commands import complete_command_aliases
from cmdbox.cli.completions.fields import (
    command_field_options,
    command_editable_field_options,
)
from cmdbox.cli.completions.tags import complete_tag_names
from cmdbox.cli.handlers import command_handlers


app = typer.Typer(no_args_is_help=True)

cli_guard = make_cli_guard(container.get_console)


@app.command("add")
@cli_guard
def add(
    alias: Annotated[
        str,
        typer.Argument(
            help="The name of the command.  Will be used to recall the command."
        ),
    ] = None,
    template: Annotated[
        str,
        typer.Argument(
            help="The actual command value that will be executed when the command is recalled using the alias"
        ),
    ] = None,
    description: Annotated[
        str, typer.Option("--description", "-d", help="A description of the command.")
    ] = None,
    tags: Annotated[
        list[str],
        typer.Option(
            "--tags",
            "-t",
            help="A list of tags to associate with the command, separated by commas.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
    interactive: Annotated[
        bool,
        typer.Option("--interactive", "-i", is_flag=True, help="Interactive mode."),
    ] = False,
) -> None:
    """
    Adds a new command with an alias, a template, description, and associated tags. The
    command can be created in interactive mode if desired or if required arguments are
    not provided.

    Args:
        alias (str): The name of the command. Will be used to recall the command.
        template (str): The actual command value that will be executed when the command is
            recalled using the alias.
        description (str): A description of the command.
        tags (list[str]): A list of tags to associate with the command, separated by commas.
        interactive (bool): Specifies whether to enable interactive mode for entering command
            details.
    """
    add_cmd_args = command_handlers.AddCommandArgs(
        alias=alias,
        template=template,
        description=description,
        tags=tags,
        interactive=interactive,
    )
    command_handlers.run_add_command(
        args=add_cmd_args,
        get_cmd_services=container.get_command_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("get")
@cli_guard
def get(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to retrieve.",
            autocompletion=complete_command_aliases,
        ),
    ],
) -> None:
    command_handlers.run_get_command(
        alias=alias,
        get_cmd_services=container.get_command_services,
        get_console=container.get_console,
    )


@app.command("update")
@cli_guard
def update(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to update.",
            autocompletion=complete_command_aliases,
        ),
    ],
    template: Annotated[
        str, typer.Option("--template", "-t", help="The new template.")
    ] = None,
    description: Annotated[
        str, typer.Option("--description", "-d", help="The new description.")
    ] = None,
    new_alias: Annotated[
        str, typer.Option("--alias", "-a", help="The new alias.")
    ] = None,
    set_: Annotated[
        list[str],
        typer.Option(
            "--set",
            "-s",
            help="A list of key=value pairs to update.",
            autocompletion=command_editable_field_options,
        ),
    ] = None,
) -> None:
    command_handlers.run_update_command(
        alias=alias,
        template=template,
        description=description,
        new_alias=new_alias,
        set_pairs=set_,
        get_cmd_services=container.get_command_services,
        get_console=container.get_console,
    )


app.command("edit", hidden=True)(update)


@app.command("list")
@cli_guard
def list_cmds(
    order: Annotated[
        str,
        typer.Option(
            "--order",
            "-o",
            help="The field to order the results by.",
            autocompletion=command_field_options,
        ),
    ] = "alias",
    tags: Annotated[
        list[str],
        typer.Option(
            "--tag",
            "-t",
            help="The tag to filter by.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
    limit: Annotated[
        int,
        typer.Option("--limit", "-l", help="The maximum number of results to return."),
    ] = 10,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults fields in settings.",
            autocompletion=command_field_options,
        ),
    ] = None,
) -> None:
    command_handlers.run_list_command(
        limit=limit,
        order=order,
        tags=tags,
        fields=fields,
        get_cmd_services=container.get_command_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_command_display_field_resolver,
    )


app.command("ls", hidden=True)(list_cmds)


@app.command("search")
@cli_guard
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        int, typer.Option(help="The maximum number of results to return.")
    ] = 10,
    search_fields: Annotated[
        list[str],
        typer.Option(
            "--in",
            "-i",
            help="The fields to search within.",
            autocompletion=command_field_options,
        ),
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults to all fields.",
            autocompletion=command_field_options,
        ),
    ] = None,
) -> None:
    command_handlers.run_search_command(
        term=term,
        limit=limit,
        search_fields=search_fields,
        fields=fields,
        get_cmd_services=container.get_command_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_command_display_field_resolver,
        get_search_field_resolver=container.get_command_search_field_resolver,
    )


app.command("find", hidden=True)(search)


@app.command("delete")
@cli_guard
def delete(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to delete.",
            autocompletion=complete_command_aliases,
        ),
    ],
) -> None:
    command_handlers.run_delete_command(
        alias=alias,
        get_cmd_services=container.get_command_services,
        get_console=container.get_console,
    )


app.command("del", hidden=True)(delete)
app.command("rm", hidden=True)(delete)
app.command("remove", hidden=True)(delete)


@app.command("tag")
@cli_guard
def add_tags(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to tag.",
            autocompletion=complete_command_aliases,
        ),
    ] = None,
    tags: Annotated[
        list[str],
        typer.Argument(
            help="The tags to add to the command.", autocompletion=complete_tag_names
        ),
    ] = None,
) -> None:
    command_handlers.run_attach_tags(
        alias=alias,
        tag_names=tags,
        get_cmd_services=container.get_command_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("untag")
@cli_guard
def remove_tags(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to untag.",
            autocompletion=complete_command_aliases,
        ),
    ] = None,
    tags: Annotated[
        list[str],
        typer.Argument(
            help="The tags to remove from the command.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
) -> None:
    command_handlers.run_detach_tags(
        alias=alias,
        tag_names=tags,
        get_cmd_services=container.get_command_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )
