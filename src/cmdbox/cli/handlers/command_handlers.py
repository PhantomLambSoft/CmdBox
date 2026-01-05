from dataclasses import dataclass
from typing import Optional, Callable, Any, Sequence, Dict

import typer

from cmdbox.cli.common.update_fields import (
    merge_fields,
    parse_set_pairs,
    filter_allowed,
)
from cmdbox.cli.handlers.common_handlers import get_tags_interactive
from cmdbox.cli.prompts.prompts import (
    prompt_for_alias,
    prompt_for_template,
    prompt_for_description,
)
from cmdbox.cli.prompts.validators import AliasValidator, TemplateValidator


@dataclass(frozen=True)
class AddCommandArgs:
    alias: Optional[str]
    template: Optional[str]
    description: Optional[str]
    tags: Optional[list[str]]
    interactive: bool = False


def run_add_command(
    *,
    args: AddCommandArgs,
    get_cmd_services: Callable[[], Any],
    get_tag_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    alias = args.alias
    template = args.template
    description = args.description
    tags = args.tags

    if args.interactive or args.alias is None:
        alias = prompt_for_alias(AliasValidator())

    if args.interactive or args.template is None:
        template = prompt_for_template(TemplateValidator())

    if args.interactive or args.description is None:
        description = prompt_for_description()

    if args.interactive or args.tags is None:
        tags = get_tags_interactive(get_tag_services())
    if not tags:
        tags = None

    cmd_service = get_cmd_services()
    cmd = cmd_service.create_command(
        alias=alias,
        template=template,
        description=description,
        tags=tags,
    )
    console = get_console()
    console.success("Command successfully created:")
    console.print_command(cmd)


def run_get_command(
    *,
    alias: str,
    get_cmd_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()
    cmd = cmd_service.get_command(alias)
    console.print_command(cmd)


def run_update_command(
    *,
    alias: str,
    template: Optional[str],
    description: Optional[str],
    new_alias: Optional[str],
    set_pairs: Optional[Sequence[str]],
    get_cmd_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    allowed = {"alias", "template", "description"}
    fields: Dict[str, Any] = {}
    if template is not None:
        fields["template"] = template
    if description is not None:
        fields["description"] = description
    if new_alias is not None:
        fields["alias"] = new_alias

    fields = merge_fields(fields, parse_set_pairs(set_pairs))
    fields = filter_allowed(fields, allowed)
    if not fields:
        raise typer.BadParameter("No fields specified to update.")
    cmd_service = get_cmd_services()
    cmd = cmd_service.get_command(alias)
    cmd_service.update_command(alias, **fields)
    console = get_console()
    console.success("Command updated successfully.")
    updated_cmd = cmd_service.get_command_by_id(cmd.id)
    console.print_command(updated_cmd)


def run_list_command(
    *,
    limit: int,
    order: str,
    tags: list[str],
    fields: list[str],
    get_cmd_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()
    cmds = cmd_service.list_commands(limit=limit, order_by=order, tags=tags)
    console.print_command_list(cmds, output_fields=fields)


def run_search_command(
    *,
    term: str,
    limit: int,
    search_fields: list[str],
    fields: list[str],
    get_cmd_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()
    cmds = cmd_service.search_commands(term, limit=limit, fields=search_fields)
    console.print_command_list(cmds, output_fields=fields)


def run_delete_command(
    *,
    alias: str,
    get_cmd_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()
    cmd = cmd_service.get_command(alias)
    if cmd_service.delete_command(alias):
        console.success("Command deleted successfully.")
        console.print_command(cmd)
    else:
        console.error(f"Failed to delete command '{alias}'.")


def run_attach_tags(
    *,
    alias: str | None = None,
    tag_names: list[str] | None = None,
    get_cmd_services: Callable[[], Any],
    get_tag_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    if alias is None:
        alias = prompt_for_alias(AliasValidator())
    if tag_names is None:
        tag_names = get_tags_interactive(get_tag_services())
    cmd_service = get_cmd_services()
    result = cmd_service.add_tags(alias=alias, tags=tag_names)
    console = get_console()
    console.print_tag_attach_result(result)


def run_detach_tags(
    *,
    alias: str | None = None,
    tag_names: list[str] | None = None,
    get_cmd_services: Callable[[], Any],
    get_tag_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    if alias is None:
        alias = prompt_for_alias(AliasValidator())
    if tag_names is None:
        tag_names = get_tags_interactive(get_tag_services())
    cmd_service = get_cmd_services()
    result = cmd_service.remove_tags(alias=alias, tags=tag_names)
    console = get_console()
    console.print_tag_detach_result(result)
