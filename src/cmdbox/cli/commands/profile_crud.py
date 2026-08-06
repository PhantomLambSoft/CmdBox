import logging
from typing import Annotated, Optional

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.completions.profiles import complete_profile_names
from cmdbox.cli.completions.fields import profile_field_options
from cmdbox.cli.handlers import profile_handler
from cmdbox.cli.commands.profile_fallback import ProfileFallback

app = typer.Typer(no_args_is_help=True, cls=ProfileFallback)

cli_guard = make_cli_guard(container.get_console)

log = logging.getLogger(__name__)


@app.command("add")
@cli_guard
def add(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the profile.", autocompletion=complete_profile_names
        ),
    ] = None,
    description: Annotated[
        str, typer.Argument(help="A description of the profile.")
    ] = None,
    interactive: Annotated[
        bool,
        typer.Option(
            "--interactive",
            "-i",
            is_flag=True,
            help="Prompt for profile details interactively.",
        ),
    ] = False,
) -> None:
    """
    Adds a new profile with a name and description.
    """
    log.debug(
        "profile.add called. name=%s, description=%s, interactive=%s",
        name,
        description,
        interactive,
    )
    add_profile_args = profile_handler.AddProfileArgs(
        name=name, description=description, interactive=interactive
    )
    profile_handler.run_add_profile(
        args=add_profile_args,
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


@app.command("get")
@cli_guard
def get(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the profile to retrieve.",
            autocompletion=complete_profile_names,
        ),
    ],
) -> None:
    """
    Gets and displays a saved profile under the provided name.
    """
    log.debug("profile.get called. name=%s", name)
    profile_handler.run_get_profile(
        name=name,
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


@app.command("update")
@cli_guard
def update(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the profile to update.",
            autocompletion=complete_profile_names,
        ),
    ],
    description: Annotated[
        str,
        typer.Option("--description", "-d", help="The new description of the profile."),
    ] = None,
    new_name: Annotated[
        str,
        typer.Option("--name", "-n", help="The new name of the profile."),
    ] = None,
    set_: Annotated[
        list[str],
        typer.Option("--set", "-s", help="A list of key=value pairs to update."),
    ] = None,
    edit_mode: Annotated[
        bool,
        typer.Option("--edit", "-e", help="Edit mode."),
    ] = False,
    edit_fields: Annotated[
        str,
        typer.Option(
            "--edit-fields",
            help="A list of fields to be edited in edit mode, separated by commas. Defaults to all fields.",
        ),
    ] = None,
) -> None:
    """
    Updates an existing profile with the provided options.
    """
    log.debug(
        "profile.update called. name=%s, description=%s, new_name=%s, set_pairs=%s, edit_mode=%s, edit_fields=%s",
        name,
        description,
        new_name,
        set_,
        edit_mode,
        edit_fields,
    )
    profile_handler.run_update_profile(
        name=name,
        description=description,
        new_name=new_name,
        set_pairs=set_,
        edit_mode=edit_mode,
        edit_fields=edit_fields,
        get_profile_services=container.get_profile_service,
        get_settings=container.get_settings,
        get_console=container.get_console,
    )


app.command("edit", hidden=True)(update)


@app.command("list")
@cli_guard
def list_profiles(
    order: Annotated[
        Optional[str],
        typer.Option("--order", "-o", help="The field to order the results by."),
    ] = None,
    limit: Annotated[
        Optional[int],
        typer.Option("--limit", "-l", help="The maximum number of results to return."),
    ] = None,
    page: Annotated[
        bool | None,
        typer.Option(
            "--page/--no-page",
            help="Force or disable the interactive pager, overriding the pager mode in settings.",
        ),
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field", "-f", help="The fields to display in the results list."
        ),
    ] = None,
) -> None:
    """
    Displays all stored profiles in a list format.
    """
    log.debug(
        "profile.list called. order=%s, limit=%s, fields=%s, page=%s",
        order,
        limit,
        fields,
        page,
    )
    profile_handler.run_list_profiles(
        limit=limit,
        page=page,
        fields=fields,
        order_by=order,
        get_profile_services=container.get_profile_service,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_profile_display_field_resolver,
    )


app.command("ls", hidden=True)(list_profiles)


@app.command("search")
@cli_guard
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        int | None, typer.Option(help="The maximum number of results to return.")
    ] = None,
    page: Annotated[
        bool | None,
        typer.Option(
            "--page/--no-page",
            help="Force or disable the interactive pager, overriding the pager mode in settings.",
        ),
    ] = None,
    search_fields: Annotated[
        list[str] | None,
        typer.Option(
            "--in",
            "-i",
            help="The fields to search within.",
            autocompletion=profile_field_options,
        ),
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--fields",
            "-f",
            help="The fields to display.",
            autocompletion=profile_field_options,
        ),
    ] = None,
) -> None:
    log.debug(
        "profile.search called. term=%s, limit=%s, search_fields=%s, fields=%s, page=%s",
        term,
        limit,
        search_fields,
        fields,
        page,
    )
    profile_handler.run_search_profiles(
        term=term,
        limit=limit,
        search_fields=search_fields,
        fields=fields,
        page=page,
        get_profile_services=container.get_profile_service,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_profile_display_field_resolver,
        get_search_field_resolver=container.get_profile_search_field_resolver,
    )


app.command("find", hidden=True)(search)


@app.command("delete")
@cli_guard
def delete(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the profile to delete.",
            autocompletion=complete_profile_names,
        ),
    ],
    force: Annotated[
        bool,
        typer.Option(
            "--force",
            help="Delete the profile even if it still has commands or variables assigned to it.",
        ),
    ] = False,
) -> None:
    """
    Deletes the profile stored under the provided name. Blocked if the profile
    still has commands or variables assigned to it, unless --force is used.
    """
    log.debug("profile.delete called. name=%s, force=%s", name, force)
    profile_handler.run_delete_profile(
        name=name,
        force=force,
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


app.command("del", hidden=True)(delete)
app.command("rm", hidden=True)(delete)
app.command("remove", hidden=True)(delete)


@app.command("status")
@cli_guard
def status() -> None:
    """
    Shows the currently active profile for commands, variables, and settings.
    """
    log.debug("profile.status called.")
    profile_handler.run_profile_status(
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


app.command("where", hidden=False)(status)


@app.command("cmd")
@cli_guard
def switch_cmd(
    name: Annotated[
        str,
        typer.Argument(
            help="The profile to activate for commands.",
            autocompletion=complete_profile_names,
        ),
    ],
) -> None:
    """
    Switches only the active command profile.
    """
    log.debug("profile.cmd called. name=%s", name)
    profile_handler.run_switch_command_profile(
        name=name,
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


@app.command("var")
@cli_guard
def switch_var(
    name: Annotated[
        str,
        typer.Argument(
            help="The profile to activate for variables.",
            autocompletion=complete_profile_names,
        ),
    ],
) -> None:
    """
    Switches only the active variable profile.
    """
    log.debug("profile.var called. name=%s", name)
    profile_handler.run_switch_variable_profile(
        name=name,
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


@app.command("settings")
@cli_guard
def switch_settings(
    name: Annotated[
        str,
        typer.Argument(
            help="The profile whose settings file should become active.",
            autocompletion=complete_profile_names,
        ),
    ],
) -> None:
    """
    Switches only the active settings profile.
    """
    log.debug("profile.settings called. name=%s", name)
    profile_handler.run_switch_settings_profile(
        name=name,
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


@app.command("switch")
@cli_guard
def switch(
    name: Annotated[
        str,
        typer.Argument(
            help="The profile to activate.", autocompletion=complete_profile_names
        ),
    ],
    cmd: Annotated[
        bool, typer.Option("--cmd", help="Only switch the command profile.")
    ] = False,
    var: Annotated[
        bool, typer.Option("--var", help="Only switch the variable profile.")
    ] = False,
    settings: Annotated[
        bool, typer.Option("--settings", help="Only switch the settings profile.")
    ] = False,
) -> None:
    """
    Switches the active profile. With no scope flags, switches command,
    variable, and settings profiles together. With one or more of
    --cmd/--var/--settings, switches only those.
    """
    log.debug(
        "profile.switch called. name=%s, cmd=%s, var=%s, settings=%s",
        name,
        cmd,
        var,
        settings,
    )
    profile_handler.run_switch_profile(
        name=name,
        cmd=cmd,
        var=var,
        settings=settings,
        get_profile_services=container.get_profile_service,
        get_console=container.get_console,
    )


app.command("set", hidden=True)(switch)
