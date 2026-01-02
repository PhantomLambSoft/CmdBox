from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.prompts.completers import TagCompleter
from cmdbox.cli.prompts.prompts import (
    prompt_for_alias,
    prompt_for_template,
    prompt_for_description,
    prompt_for_tags,
)
from cmdbox.cli.prompts.validators import (
    AliasValidator,
    TemplateValidator,
    TagNameValidator,
)
from cmdbox.repositories.errors import UnknownAliasError

app = typer.Typer(no_args_is_help=True)


@app.command("add")
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
        ),
    ] = None,
    interactive: Annotated[
        bool, typer.Option("--interactive", "-i", help="Interactive mode.")
    ] = False,
):
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
    if interactive or alias is None:
        alias_validator = AliasValidator()
        alias = prompt_for_alias(alias_validator)
    if interactive or template is None:
        template_validator = TemplateValidator()
        template = prompt_for_template(template_validator)
    if interactive or description is None:
        description = prompt_for_description()
    if interactive or tags is None:
        tag_services = container.get_tag_services()

        def get_tags(query: str) -> list[str]:
            found_tags = tag_services.search(query, fields="name")
            return [tag.name for tag in found_tags]

        tag_completer = TagCompleter(get_tags)
        tag_validator = TagNameValidator()
        tags = prompt_for_tags(tag_completer, tag_validator)
    if not tags:
        tags = None
    cmd_service = container.get_command_services()
    cmd = cmd_service.create_command(
        alias=alias,
        template=template,
        description=description,
        tags=tags,
    )
    console = container.get_console()
    console.success("Command successfully created:")
    console.print_command(cmd)


@app.command("get")
def get(
    alias: Annotated[
        str,
        typer.Argument(help="The alias of the command to retrieve."),
    ],
):
    console = container.get_console()
    cmd_service = container.get_command_services()
    try:
        cmd = cmd_service.get_command(alias)
        console.print_command(cmd)
    except UnknownAliasError:
        console.error(f"Command '{alias}' not found.")


@app.command("update")
def update():
    console = container.get_console()
    console.error("Not yet implemented.")


@app.command("list")
def list_cmds(
    limit: Annotated[
        int, typer.Option(help="The maximum number of results to return.")
    ] = 10,
):
    console = container.get_console()
    cmd_service = container.get_command_services()
    cmds = cmd_service.list_commands(limit=limit)
    console.print_command_list(cmds)


@app.command("search")
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        int, typer.Option(help="The maximum number of results to return.")
    ] = 10,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to search within. Defaults to all fields.",
        ),
    ] = None,
):
    console = container.get_console()
    cmd_service = container.get_command_services()
    if fields is None:
        fields = ["alias", "template", "description"]
    cmds = cmd_service.search(term, fields=fields, limit=limit)
    console.print_command_list(cmds)


@app.command("delete")
def delete(
    alias: Annotated[
        str,
        typer.Argument(help="The alias of the command to delete."),
    ],
):
    console = container.get_console()
    cmd_service = container.get_command_services()
    if cmd_service.delete_command(alias):
        console.success(f"Command '{alias}' deleted successfully.")
    else:
        console.error(f"Command '{alias}' not found.")
