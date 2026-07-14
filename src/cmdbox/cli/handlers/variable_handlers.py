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
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.presenters.variable_presenter import (
    render_variable_created,
    render_variable,
    render_variable_list,
    render_variable_updated,
    render_variable_deleted,
)
from cmdbox.cli.ui.presenters.result_presenter import (
    render_tag_attach_result,
    render_tag_detach_result,
)
from cmdbox.services.field_selection import FieldSelectionResolver
from cmdbox.services.variable_services import VariableServices
from cmdbox.services.tag_services import TagServices
from cmdbox.settings.models import Settings
from cmdbox.logging_setup.log_decorators import log_action


@dataclass(frozen=True)
class AddVariableArgs:
    name: Optional[str]
    value: Optional[str]
    tags: Optional[list[str]]
    interactive: bool = False


@log_action(__name__, "run_add_variable")
def run_add_variable(
    *,
    args: AddVariableArgs,
    get_var_services: Callable[[], VariableServices],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    name = args.name
    value = args.value
    tags = args.tags

    if args.interactive or name is None:
        name = prompt_for_name(NameValidator())

    if args.interactive or value is None:
        value = prompt_for_value()

    if args.interactive or tags is None:
        tags = get_tags_interactive(get_tag_services())
    if not tags:
        tags = None

    var_service = get_var_services()
    var = var_service.create_variable(name=name, value=value, tags=tags)
    console = get_console()
    console.print(render_variable_created(var))


@log_action(__name__, "run_get_variable")
def run_get_variable(
    *,
    name: str,
    get_var_services: Callable[[], VariableServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    var_service = get_var_services()
    var = var_service.get_variable(name)
    rendered_var = render_variable(var)
    console.print(rendered_var)


@log_action(__name__, "run_update_variable")
def run_update_variable(
    *,
    name: str,
    value: Optional[str],
    new_name: Optional[str],
    set_pairs: Optional[Sequence[str]],
    edit_mode: bool,
    edit_fields: Optional[str],
    get_var_services: Callable[[], VariableServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
) -> None:
    allowed = {"name", "value"}
    fields: Dict[str, Any] = {}

    var_service = get_var_services()
    var = var_service.get_variable(name)
    console = get_console()

    if edit_mode:
        if any([new_name, value, set_pairs]):
            raise typer.BadParameter(
                "--edit cannot be combined with field options or --set."
            )

        updated_fields: dict[str, Any] = {}
        if edit_fields:
            edit_fields = [x.strip() for x in edit_fields.split(",")]

        field_aliases = get_settings().field_aliases.alias_mapping

        def check_field_aliases(field: str) -> bool:
            return (
                edit_fields is None
                or field in edit_fields
                or any(x in edit_fields for x in field_aliases.get(field, []))
            )

        if check_field_aliases("name"):
            updated_fields["name"] = prompt_for_name(NameValidator(), default=var.name)
        if check_field_aliases("value"):
            updated_fields["value"] = prompt_for_value(default=var.value)

        fields = merge_fields(fields, updated_fields)

    else:
        if value is not None:
            fields["value"] = value
        if new_name is not None:
            fields["name"] = new_name

        fields = merge_fields(fields, parse_set_pairs(set_pairs))
        fields = filter_allowed(fields, allowed)

        if not fields:
            raise typer.BadParameter("No fields specified to update.")

    current = {
        "name": var.name,
        "value": var.value,
    }
    fields = {key: value for key, value in fields.items() if current.get(key) != value}

    if not fields:
        console.info("No changes detected.")
        return

    var_service.update_variable(name, **fields)

    updated_var = var_service.get_variable_by_id(var.id)
    console.print(render_variable_updated(updated_var))


@log_action(__name__, "run_list_variables")
def run_list_variables(
    *,
    limit: int | None,
    page: bool | None,
    order_by: str | None,
    tags: list[str] | None,
    fields: list[str] | None = None,
    get_var_services: Callable[[], VariableServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
    get_display_field_resolver: Callable[[], FieldSelectionResolver],
) -> None:
    console = get_console()
    var_service = get_var_services()
    settings = get_settings()

    if limit is None:
        limit = settings.default_fields.variable_list_limit
    if order_by is None:
        order_by = settings.default_fields.variable_default_order

    vars_ = var_service.list_variables(limit=limit, order_by=order_by, tags=tags)

    fields = get_display_field_resolver().resolve(
        fields,
        default_fields=settings.default_fields.variable_output,
        aliases=settings.field_aliases.alias_map,
    )

    rendered_var_list = render_variable_list(vars_, title="Variables", fields=fields)
    console.print_paged(rendered_var_list, row_count=len(vars_), force=page)


@log_action(__name__, "run_search_variables")
def run_search_variables(
    *,
    term: str,
    limit: int,
    page: bool | None,
    search_fields: list[str] | None = None,
    fields: list[str] | None = None,
    get_var_services: Callable[[], VariableServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
    get_display_field_resolver: Callable[[], FieldSelectionResolver],
    get_search_field_resolver: Callable[[], FieldSelectionResolver],
) -> None:
    console = get_console()
    var_service = get_var_services()

    settings = get_settings()
    output_fields = get_display_field_resolver().resolve(
        fields,
        default_fields=settings.default_fields.variable_output,
        aliases=settings.field_aliases.alias_map,
    )
    search_fields = get_search_field_resolver().resolve(
        search_fields,
        default_fields=settings.default_fields.variable_search,
        aliases=settings.field_aliases.alias_map,
    )

    if limit is None:
        limit = settings.default_fields.variable_list_limit
    vars_ = var_service.search(term, limit=limit, fields=search_fields)
    rendered_var_list = render_variable_list(
        vars_, title="Search Results", fields=output_fields
    )
    console.print_paged(rendered_var_list, row_count=len(vars_), force=page)


@log_action(__name__, "run_delete_variable")
def run_delete_variable(
    *,
    name: str,
    get_var_services: Callable[[], VariableServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    var_service = get_var_services()
    var = var_service.get_variable(name)
    if var_service.delete_variable(name):
        console.print(render_variable_deleted(var))
    else:
        console.error(f"Failed to delete variable '{name}'.")


@log_action(__name__, "run_attach_tags")
def run_attach_tags(
    *,
    name: str | None = None,
    tag_names: list[str] | None = None,
    get_var_services: Callable[[], VariableServices],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    if not name:
        name = prompt_for_name(NameValidator())
    if not tag_names:
        tag_names = get_tags_interactive(get_tag_services())
    var_service = get_var_services()
    result = var_service.add_tags(name=name, tags=tag_names)
    console = get_console()
    console.print(render_tag_attach_result(result))


@log_action(__name__, "run_detach_tags")
def run_detach_tags(
    *,
    name: str | None = None,
    tag_names: list[str] | None = None,
    get_var_services: Callable[[], VariableServices],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    if not name:
        name = prompt_for_name(NameValidator())
    if not tag_names:
        tag_names = get_tags_interactive(get_tag_services())
    var_service = get_var_services()
    result = var_service.remove_tags(name=name, tags=tag_names)
    console = get_console()
    console.print(render_tag_detach_result(result))
