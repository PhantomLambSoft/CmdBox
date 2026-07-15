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
from cmdbox.cli.ui.presenters.tag_presenter import (
    render_tag_created,
    render_tag,
    render_tag_list,
    render_tag_updated,
    render_tag_deleted,
)
from cmdbox.cli.prompts.validators import TagNameValidator
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.services.field_selection import FieldSelectionResolver
from cmdbox.services.tag_services import TagServices
from cmdbox.settings.models import Settings
from cmdbox.logging_setup.log_decorators import log_action


@dataclass(frozen=True)
class AddTagArgs:
    name: Optional[str]
    description: Optional[str]
    interactive: bool = False


@log_action(__name__, "run_add_tag")
def run_add_tag(
    *,
    args: AddTagArgs,
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    name = args.name
    description = args.description

    if args.interactive or name is None:
        name = prompt_for_name(TagNameValidator())

    if args.interactive or description is None:
        description = prompt_for_description()

    tag_service = get_tag_services()
    tag = tag_service.create_tag(name=name, description=description)
    console = get_console()
    console.print(render_tag_created(tag))


@log_action(__name__, "run_get_tag")
def run_get_tag(
    *,
    name: str,
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    tag_service = get_tag_services()
    tag = tag_service.get_tag(name)
    console.print(render_tag(tag))


@log_action(__name__, "run_update_tag")
def run_update_tag(
    *,
    name: str,
    description: Optional[str],
    new_name: Optional[str],
    set_pairs: Optional[Sequence[str]],
    edit_mode: bool,
    edit_fields: Optional[str],
    get_tag_services: Callable[[], TagServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
) -> None:
    allowed = {"name", "description"}
    fields: Dict[str, Any] = {}

    tag_service = get_tag_services()
    tag = tag_service.get_tag(name)
    console = get_console()

    if edit_mode:
        if any([description, new_name, set_pairs]):
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
            updated_fields["name"] = prompt_for_name(
                TagNameValidator(), default=tag.name
            )
        if check_field_aliases("description"):
            updated_fields["description"] = prompt_for_description(
                default=tag.description
            )

        fields = updated_fields

    else:
        if description is not None:
            fields["description"] = description
        if new_name is not None:
            fields["name"] = new_name

        fields = merge_fields(fields, parse_set_pairs(set_pairs))
        fields = filter_allowed(fields, allowed)

        if not fields:
            raise typer.BadParameter("No fields specified to update.")

    current = {
        "name": tag.name,
        "description": tag.description,
    }
    fields = {key: value for key, value in fields.items() if current.get(key) != value}

    if not fields:
        console.info("No changes detected.")
        return

    tag_service.update_tag(name, **fields)

    updated_tag = tag_service.get_tag_by_id(tag.id)
    console.print(render_tag_updated(updated_tag))


@log_action(__name__, "run_list_tags")
def run_list_tags(
    *,
    limit: int | None,
    page: bool | None,
    order_by: str | None,
    fields: list[str] | None = None,
    get_tag_services: Callable[[], TagServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
    get_display_field_resolver: Callable[[], FieldSelectionResolver],
) -> None:
    console = get_console()
    tag_service = get_tag_services()
    settings = get_settings()

    if limit is None:
        limit = settings.default_fields.tag_list_limit
    if order_by is None:
        order_by = settings.default_fields.tag_default_order

    tags = tag_service.list_tags(limit=limit, order_by=order_by)

    fields = get_display_field_resolver().resolve(
        fields,
        default_fields=settings.default_fields.tag_output,
        aliases=settings.field_aliases.alias_map,
    )

    rendered_tag_list = render_tag_list(tags, title="Tags", fields=fields)
    console.print_paged(rendered_tag_list, row_count=len(tags), force=page)


@log_action(__name__, "run_search_tags")
def run_search_tags(
    *,
    term: str,
    limit: int | None,
    page: bool | None,
    search_fields: list[str] | None = None,
    fields: list[str] | None = None,
    get_tag_services: Callable[[], TagServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
    get_display_field_resolver: Callable[[], FieldSelectionResolver],
    get_search_field_resolver: Callable[[], FieldSelectionResolver],
) -> None:
    console = get_console()
    tag_service = get_tag_services()
    settings = get_settings()

    output_fields = get_display_field_resolver().resolve(
        fields,
        default_fields=settings.default_fields.tag_output,
        aliases=settings.field_aliases.alias_map,
    )
    search_fields = get_search_field_resolver().resolve(
        search_fields,
        default_fields=settings.default_fields.tag_search,
        aliases=settings.field_aliases.alias_map,
    )

    if limit is None:
        limit = settings.default_fields.tag_list_limit
    tags = tag_service.search(term, limit=limit, fields=search_fields)
    rendered_tag_list = render_tag_list(
        tags, title="Search Results", fields=output_fields
    )
    console.print_paged(rendered_tag_list, row_count=len(tags), force=page)


@log_action(__name__, "run_delete_tag")
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
        console.print(render_tag_deleted(tag))
    else:
        console.error(f"Failed to delete tag '{name}'.")
