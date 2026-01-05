from dataclasses import dataclass
from typing import Optional, Callable, Any, Sequence, Dict

import typer

from cmdbox.cli.common.update_fields import (
    merge_fields,
    parse_set_pairs,
    filter_allowed,
)
from cmdbox.cli.prompts.prompts import (
    prompt_for_name,
    prompt_for_description,
)
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.services.tag_services import TagServices


@dataclass(frozen=True)
class AddTagArgs:
    name: Optional[str]
    description: Optional[str]
    interactive: bool = False


def run_add_tag(
    *,
    args: AddTagArgs,
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    name = args.name
    description = args.description

    if args.interactive or name is None:
        name = prompt_for_name()

    if args.interactive or description is None:
        description = prompt_for_description()

    tag_service = get_tag_services()
    tag = tag_service.create_tag(name=name, description=description)
    console = get_console()
    console.success("Tag successfully created:")
    console.print_tag(tag)


def run_get_tag(
    *,
    name: str,
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    tag_service = get_tag_services()
    tag = tag_service.get_tag(name)
    console.print_tag(tag)


def run_update_tag(
    *,
    name: str,
    description: Optional[str],
    new_name: Optional[str],
    set_pairs: Sequence[str],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    allowed = {"name", "description"}
    fields: Dict[str, Any] = {}
    if description is not None:
        fields["description"] = description
    if new_name is not None:
        fields["name"] = new_name

    fields = merge_fields(fields, parse_set_pairs(set_pairs))
    fields = filter_allowed(fields, allowed)
    if not fields:
        raise typer.BadParameter("No fields specified to update.")

    tag_service = get_tag_services()
    tag = tag_service.get_tag(name)
    tag_service.update_tag(name, **fields)
    console = get_console()
    console.success("Tag updated successfully.")
    updated_tag = tag_service.get_tag_by_id(tag.id)
    console.print_tag(updated_tag)


def run_list_tags(
    *,
    limit: int,
    order_by: str,
    fields: list[str],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    tag_service = get_tag_services()
    tags = tag_service.list_tags(limit=limit, order_by=order_by)
    console.print_tag_list(tags, output_fields=fields)


def run_search_tags(
    *,
    term: str,
    limit: int,
    search_fields: list[str],
    fields: list[str],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    tag_service = get_tag_services()
    tags = tag_service.search_tags(term, limit=limit, fields=search_fields)
    console.print_tag_list(tags, output_fields=fields)


def run_delete_tag(
    *,
    name: str,
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    tag_service = get_tag_services()
    tag = tag_service.get_tag(name)
    if tag_service.delete_tag(name):
        console.success(f"Tag '{name}' deleted successfully.")
        console.print_tag(tag)
    else:
        console.error(f"Failed to delete tag '{name}'.")
