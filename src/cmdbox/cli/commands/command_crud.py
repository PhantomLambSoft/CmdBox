from typing import Annotated, Optional, List, Dict, Any

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
        bool,
        typer.Option("--interactive", "-i", is_flag=True, help="Interactive mode."),
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


def _parse_set_pairs(pairs: Optional[List[str]]) -> Dict[str, str]:
    out: Dict[str, str] = {}
    if not pairs:
        return out
    for item in pairs:
        if "=" not in item:
            raise typer.BadParameter(f"Invalid --set value '{item}'. Use key=value.")
        k, v = item.split("=", 1)
        k = k.strip()
        if not k:
            raise typer.BadParameter(
                f"Invalid --set value '{item}'. Key cannot be empty."
            )
        out[k] = v
    return out


def _merge_fields(base: Dict[str, Any], extra: Dict[str, Any]) -> Dict[str, Any]:
    conflicts = set(base).intersection(extra)
    if conflicts:
        keys = ", ".join(sorted(conflicts))
        raise typer.BadParameter(f"Field(s) specified multiple ways: {keys}")
    merged = dict(base)
    merged.update(extra)
    return merged


@app.command("update")
def update(
    alias: Annotated[str, typer.Argument(help="The alias of the command to update.")],
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
        List[str],
        typer.Option("--set", "-s", help="A list of key=value pairs to update."),
    ] = None,
):
    fields: Dict[str, Any] = {}
    if template is not None:
        fields["template"] = template
    if description is not None:
        fields["description"] = description
    if new_alias is not None:
        fields["alias"] = new_alias

    set_fields = _parse_set_pairs(set_)
    fields = _merge_fields(fields, set_fields)
    if not fields:
        raise typer.BadParameter("No fields specified to update.")

    cmd_service = container.get_command_services()

    cmd = cmd_service.get_command(alias)

    cmd_service.update_command(alias, **fields)
    console = container.get_console()
    console.success("Command updated successfully.")
    updated_cmd = cmd_service.get_command_by_id(cmd.id)
    console.print_command(updated_cmd)


@app.command("list")
def list_cmds(
    order: Annotated[
        str, typer.Option("--order", "-o", help="The field to order the results by.")
    ] = "alias",
    tags: Annotated[
        list[str], typer.Option("--tag", "-t", help="The tag to filter by.")
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
            help="The field(s) to display in the results list. Defaults to all fields.",
        ),
    ] = None,
):
    console = container.get_console()
    cmd_service = container.get_command_services()
    cmds = cmd_service.list_commands(order_by=order, tags=tags, limit=limit)
    console.print_command_list(cmds, output_fields=fields)


@app.command("search")
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        int, typer.Option(help="The maximum number of results to return.")
    ] = 10,
    search_fields: Annotated[
        list[str], typer.Option("--in", "-i", help="The fields to search within.")
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults to all fields.",
        ),
    ] = None,
):
    console = container.get_console()
    cmd_service = container.get_command_services()
    cmds = cmd_service.search(term, fields=search_fields, limit=limit)
    console.print_command_list(cmds, output_fields=fields)


@app.command("delete")
def delete(
    alias: Annotated[
        str,
        typer.Argument(help="The alias of the command to delete."),
    ],
):
    console = container.get_console()
    cmd_service = container.get_command_services()
    cmd = cmd_service.get_command(alias)
    if cmd_service.delete_command(alias):
        console.success("Command deleted successfully.")
        console.print_command(cmd)
    else:
        console.error(f"Command '{alias}' not found.")
