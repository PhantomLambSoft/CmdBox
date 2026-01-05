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
    prompt_for_name,
    prompt_for_value,
)
from cmdbox.cli.prompts.validators import NameValidator


@dataclass(frozen=True)
class AddVariableArgs:
    name: Optional[str]
    value: Optional[str]
    tags: Optional[list[str]]
    interactive: bool = False


def run_add_variable(
    *,
    args: AddVariableArgs,
    get_var_services: Callable[[], Any],
    get_tag_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    name = args.name
    value = args.value
    tags = args.tags

    if args.interactive or name is None:
        name = prompt_for_name()
    if args.interactive or value is None:
        value = prompt_for_value()
    if args.interactive or tags is None:
        tags = get_tags_interactive(get_tag_services())
    if args.interactive or not tags:
        tags = get_tags_interactive(get_tag_services())
    if not tags:
        tags = None

    var_service = get_var_services()
    var = var_service.create_variable(name=name, value=value, tags=tags)
    console = get_console()
    console.success("Variable successfully created:")
    console.print_variable(var)


def run_get_variable(
    *,
    name: str,
    get_var_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    var_service = get_var_services()
    var = var_service.get_variable(name)
    console.print_variable(var)


def run_update_variable(
    *,
    name: str,
    value: Optional[str],
    new_name: Optional[str],
    set_pairs: Optional[Sequence[str]],
    get_var_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    allowed = {"name", "value"}
    fields: Dict[str, Any] = {}
    if value is not None:
        fields["value"] = value
    if new_name is not None:
        fields["name"] = new_name

    fields = merge_fields(fields, parse_set_pairs(set_pairs))
    fields = filter_allowed(fields, allowed)
    if not fields:
        raise typer.BadParameter("No fields specified to update.")
    var_service = get_var_services()
    var = var_service.get_variable(name)
    var_service.update_variable(name, **fields)
    console = get_console()
    console.success("Variable updated successfully.")
    updated_var = var_service.get_variable_by_id(var.id)
    console.print_variable(updated_var)


def run_list_variables(
    *,
    limit: int,
    order_by: str,
    tags: list[str],
    fields: list[str],
    get_var_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    var_service = get_var_services()
    vars_ = var_service.list_variables(limit=limit, order_by=order_by, tags=tags)
    console.print_variable_list(vars_, output_fields=fields)


def run_search_variables(
    *,
    term: str,
    limit: int,
    search_fields: list[str],
    fields: list[str],
    get_var_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    var_service = get_var_services()
    vars_ = var_service.search_variables(term, limit=limit, fields=search_fields)
    console.print_variable_list(vars_, output_fields=fields)


def run_delete_variable(
    *,
    name: str,
    get_var_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    console = get_console()
    var_service = get_var_services()
    var = var_service.get_variable(name)
    if var_service.delete_variable(name):
        console.success(f"Variable '{name}' deleted successfully.")
        console.print_variable(var)
    else:
        console.error(f"Failed to delete variable '{name}'.")


def run_attach_tags(
    *,
    name: str | None = None,
    tag_names: list[str] | None = None,
    get_var_services: Callable[[], Any],
    get_tag_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    if not name:
        name = prompt_for_name(NameValidator())
    if not tag_names:
        tag_names = get_tags_interactive(get_tag_services())
    var_service = get_var_services()
    result = var_service.add_tags(name=name, tags=tag_names)
    console = get_console()
    console.print_tag_attach_result(result)


def run_detach_tags(
    *,
    name: str | None = None,
    tag_names: list[str] | None = None,
    get_var_services: Callable[[], Any],
    get_tag_services: Callable[[], Any],
    get_console: Callable[[], Any],
) -> None:
    if not name:
        name = prompt_for_name(NameValidator())
    if not tag_names:
        tag_names = get_tags_interactive(get_tag_services())
    var_service = get_var_services()
    result = var_service.remove_tags(name=name, tags=tag_names)
    console = get_console()
    console.print_tag_detach_result(result)
