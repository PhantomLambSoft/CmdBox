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
from cmdbox.cli.ui.presenters.profile_presenter import (
    render_profile_created,
    render_profile,
    render_profile_list,
    render_profile_updated,
    render_profile_deleted,
    render_profile_switched,
    render_profile_status,
)
from cmdbox.cli.prompts.validators import NameValidator
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.services.profile_services import ProfileServices
from cmdbox.settings.models import Settings
from cmdbox.logging_setup.log_decorators import log_action


@dataclass(frozen=True)
class AddProfileArgs:
    name: Optional[str] = None
    description: Optional[str] = None
    interactive: bool = False


@log_action(__name__, "run_add_profile")
def run_add_profile(
    *,
    args: AddProfileArgs,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    name = args.name
    description = args.description

    if args.interactive or name is None:
        name = prompt_for_name(NameValidator())

    if args.interactive or description is None:
        description = prompt_for_description()

    profile_service = get_profile_services()
    profile = profile_service.create_profile(name=name, description=description)
    console = get_console()
    console.print(render_profile_created(profile))


@log_action(__name__, "run_get_profile")
def run_get_profile(
    *,
    name: str,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    profile_service = get_profile_services()
    profile = profile_service.get_profile(name=name)
    console.print(render_profile(profile))


@log_action(__name__, "run_update_profile")
def run_update_profile(
    *,
    name: str,
    description: Optional[str],
    new_name: Optional[str],
    set_pairs: Optional[Sequence[str]],
    edit_mode: bool,
    edit_fields: Optional[str],
    get_profile_services: Callable[[], ProfileServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
) -> None:
    allowed = {"name", "description"}
    fields: Dict[str, Any] = {}

    profile_service = get_profile_services()
    profile = profile_service.get_profile(name)
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
                NameValidator(), default=profile.name
            )
        if check_field_aliases("description"):
            updated_fields["description"] = prompt_for_description(
                default=profile.description
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
        "name": profile.name,
        "description": profile.description,
    }
    fields = {key: value for key, value in fields.items() if current.get(key) != value}

    if not fields:
        console.info("No changes detected.")
        return

    profile_service.update_profile(name, **fields)

    updated_profile = profile_service.get_profile(fields.get("name", name))
    console.print(render_profile_updated(updated_profile))


@log_action(__name__, "run_list_profiles")
def run_list_profiles(
    *,
    limit: int | None,
    page: bool | None,
    order_by: str | None,
    fields: list[str] | None = None,
    get_profile_services: Callable[[], ProfileServices],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    profile_service = get_profile_services()
    settings = get_settings()

    if limit is None:
        limit = settings.default_fields.tag_list_limit  # see note below
    if order_by is None:
        order_by = "name"

    profiles = profile_service.list_profiles(limit=limit, order_by=order_by)

    rendered_profile_list = render_profile_list(
        profiles, title="Profiles", fields=fields
    )
    console.print_paged(rendered_profile_list, row_count=len(profiles), force=page)


@log_action(__name__, "run_delete_profile")
def run_delete_profile(
    *,
    name: str,
    force: bool,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    profile_service = get_profile_services()
    profile = profile_service.get_profile(name)
    if profile_service.delete_profile(name, force=force):
        console.print(render_profile_deleted(profile))
    else:
        console.error(f"Failed to delete profile '{name}'.")


@log_action(__name__, "run_switch_command_profile")
def run_switch_command_profile(
    *,
    name: str,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    profile_service = get_profile_services()
    profile_service.switch_command_profile(name)
    profile = profile_service.get_profile(name)
    console.print(render_profile_switched(profile, scope="command"))


@log_action(__name__, "run_switch_variable_profile")
def run_switch_variable_profile(
    *,
    name: str,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    profile_service = get_profile_services()
    profile_service.switch_variable_profile(name)
    profile = profile_service.get_profile(name)
    console.print(render_profile_switched(profile, scope="variable"))


@log_action(__name__, "run_switch_settings_profile")
def run_switch_settings_profile(
    *,
    name: str,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    profile_service = get_profile_services()
    profile_service.switch_settings_profile(name)
    profile = profile_service.get_profile(name)
    console.print(render_profile_switched(profile, scope="settings"))


@log_action(__name__, "run_switch_profile")
def run_switch_profile(
    *,
    name: str,
    cmd: bool,
    var: bool,
    settings: bool,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    """
    Switches the active profile. With no scope flags, switches command,
    variable, and settings profiles all at once (the linked default). With
    one or more of --cmd/--var/--settings, switches only those.
    """
    console = get_console()
    profile_service = get_profile_services()

    if not (cmd or var or settings):
        profile_service.switch_profile(name)
        profile = profile_service.get_profile(name)
        console.print(
            render_profile_switched(profile, scope="command, variable, and settings")
        )
        return

    switched = []
    if cmd:
        profile_service.switch_command_profile(name)
        switched.append("command")
    if var:
        profile_service.switch_variable_profile(name)
        switched.append("variable")
    if settings:
        profile_service.switch_settings_profile(name)
        switched.append("settings")

    profile = profile_service.get_profile(name)
    console.print(render_profile_switched(profile, scope=", ".join(switched)))


@log_action(__name__, "run_profile_status")
def run_profile_status(
    *,
    get_profile_services: Callable[[], ProfileServices],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    profile_service = get_profile_services()
    status = profile_service.get_status()
    console.print(render_profile_status(status))
