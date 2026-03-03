from dataclasses import dataclass
from typing import Optional, Callable, Any, Sequence

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
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.presenters.command_presenter import (
    render_command,
    render_command_list,
    render_command_created,
    render_command_updated,
    render_command_deleted,
)
from cmdbox.cli.ui.presenters.result_presenter import (
    render_tag_attach_result,
    render_tag_detach_result,
)
from cmdbox.services.command_services import CommandServices
from cmdbox.services.field_selection import FieldSelectionResolver
from cmdbox.services.tag_services import TagServices
from cmdbox.settings.models import Settings
from cmdbox.logging_setup.log_decorators import log_action


@dataclass(frozen=True)
class AddCommandArgs:
    alias: Optional[str]
    template: Optional[str]
    description: Optional[str]
    tags: Optional[list[str]]
    interactive: bool = False


@log_action(__name__, "run_add_command")
def run_add_command(
    *,
    args: AddCommandArgs,
    get_cmd_services: Callable[[], CommandServices],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
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
    console.print(render_command_created(cmd))


@log_action(__name__, "run_get_command")
def run_get_command(
    *,
    alias: str,
    get_cmd_services: Callable[[], CommandServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()
    cmd = cmd_service.get_command(alias)
    rendered_cmd = render_command(cmd)
    console.print(rendered_cmd)


@log_action(__name__, "run_update_command")
def run_update_command(
    *,
    alias: str,
    template: Optional[str],
    description: Optional[str],
    new_alias: Optional[str],
    set_pairs: Optional[Sequence[str]],
    edit_mode: bool,
    edit_fields: Optional[str],
    get_cmd_services: Callable[[], CommandServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
) -> None:
    allowed = {"alias", "template", "description"}
    fields: dict[str, Any] = {}

    cmd_service = get_cmd_services()
    cmd = cmd_service.get_command(alias)
    console = get_console()

    if edit_mode:
        if any([template, description, new_alias, set_pairs]):
            raise typer.BadParameter(
                "--edit cannot be combined with field options or --set"
            )

        updated_fields: dict[str, Any] = {}
        if edit_fields:
            edit_fields = [x.strip() for x in edit_fields.split(",")]

        field_aliases = get_settings().field_aliases.alias_mapping

        def check_field_alias(field: str) -> bool:
            return (
                edit_fields is None
                or field in edit_fields
                or any(x in edit_fields for x in field_aliases.get(field, []))
            )

        if check_field_alias("alias"):
            updated_fields["alias"] = prompt_for_alias(
                AliasValidator(), default=cmd.alias
            )
        if check_field_alias("template"):
            updated_fields["template"] = prompt_for_template(
                TemplateValidator(), default=cmd.template
            )
        if check_field_alias("description"):
            updated_fields["description"] = prompt_for_description(
                default=cmd.description
            )

        fields = updated_fields

    else:
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

    current = {
        "alias": cmd.alias,
        "template": cmd.template,
        "description": cmd.description,
    }
    fields = {key: value for key, value in fields.items() if current.get(key) != value}

    if not fields:
        console.info("No changes detected.")
        return

    cmd_service.update_command(alias, **fields)

    updated_cmd = cmd_service.get_command_by_id(cmd.id)
    console.print(render_command_updated(updated_cmd))


@log_action(__name__, "run_list_command")
def run_list_command(
    *,
    limit: int,
    order: str,
    tags: list[str],
    fields: list[str] | None = None,
    get_cmd_services: Callable[[], CommandServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
    get_display_field_resolver: Callable[[], FieldSelectionResolver],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()

    settings = get_settings()
    resolved_fields = get_display_field_resolver().resolve(
        fields,
        default_fields=settings.default_fields.command_output,
        aliases=settings.field_aliases.alias_map,
    )

    cmds = cmd_service.list_commands(limit=limit, order_by=order, tags=tags)
    rendered_cmd_list = render_command_list(
        cmds, title="Commands", fields=resolved_fields
    )
    console.print(rendered_cmd_list)


@log_action(__name__, "run_search_command")
def run_search_command(
    *,
    term: str,
    limit: int,
    search_fields: list[str] | None = None,
    fields: list[str] | None = None,
    get_cmd_services: Callable[[], CommandServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
    get_display_field_resolver: Callable[[], FieldSelectionResolver],
    get_search_field_resolver: Callable[[], FieldSelectionResolver],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()

    settings = get_settings()
    output_fields = get_display_field_resolver().resolve(
        fields,
        default_fields=settings.default_fields.command_output,
        aliases=settings.field_aliases.alias_map,
    )
    search_fields = get_search_field_resolver().resolve(
        search_fields,
        default_fields=settings.default_fields.command_search,
        aliases=settings.field_aliases.alias_map,
    )

    cmds = cmd_service.search(term, limit=limit, fields=search_fields)
    rendered_cmd_list = render_command_list(
        cmds, title="Search Results", fields=output_fields
    )
    console.print(rendered_cmd_list)


@log_action(__name__, "run_delete_command")
def run_delete_command(
    *,
    alias: str,
    get_cmd_services: Callable[[], CommandServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    cmd_service = get_cmd_services()
    cmd = cmd_service.get_command(alias)
    if cmd_service.delete_command(alias):
        console.print(render_command_deleted(cmd))
    else:
        console.error(f"Failed to delete command '{alias}'.")


@log_action(__name__, "run_attach_tags")
def run_attach_tags(
    *,
    alias: str | None = None,
    tag_names: list[str] | None = None,
    get_cmd_services: Callable[[], CommandServices],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    if alias is None:
        alias = prompt_for_alias(AliasValidator())
    if tag_names is None:
        tag_names = get_tags_interactive(get_tag_services())
    cmd_service = get_cmd_services()
    result = cmd_service.add_tags(alias=alias, tags=tag_names)
    console = get_console()
    console.print(render_tag_attach_result(result))


@log_action(__name__, "run_detach_tags")
def run_detach_tags(
    *,
    alias: str | None = None,
    tag_names: list[str] | None = None,
    get_cmd_services: Callable[[], CommandServices],
    get_tag_services: Callable[[], TagServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    if alias is None:
        alias = prompt_for_alias(AliasValidator())
    if tag_names is None:
        tag_names = get_tags_interactive(get_tag_services())
    cmd_service = get_cmd_services()
    result = cmd_service.remove_tags(alias=alias, tags=tag_names)
    console = get_console()
    console.print(render_tag_detach_result(result))
