from idlelib import autocomplete
from typing import Annotated, Optional
import logging

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.completions.commands import complete_command_aliases
from cmdbox.cli.completions.fields import (
    command_field_options,
    command_editable_field_options,
)
from cmdbox.cli.completions.profiles import complete_profile_names
from cmdbox.cli.completions.tags import complete_tag_names
from cmdbox.cli.handlers import command_handlers

app = typer.Typer(no_args_is_help=True)

cli_guard = make_cli_guard(container.get_console)

log = logging.getLogger(__name__)


PROFILE_OPTION = Annotated[
    Optional[str],
    typer.Option(
        "--profile",
        "-p",
        help="The profile to target for this command. Defaults to the currently active profile.",
        autocompletion=complete_profile_names,
    ),
]


@app.command("add")
@cli_guard
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
            help="The actual command value that will be executed when the command is recalled using the alias."
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
            autocompletion=complete_tag_names,
        ),
    ] = None,
    cwd: Annotated[
        str,
        typer.Option("--cwd", "-c", help="Working directory to run the command from."),
    ] = None,
    shell: Annotated[
        str,
        typer.Option("--shell", "-s", help="Shell to use when running the command."),
    ] = None,
    env: Annotated[
        list[str],
        typer.Option(
            "--env",
            "-e",
            help="Environment variable to set when running the command, in KEY=VALUE format.",
        ),
    ] = None,
    timeout: Annotated[
        int,
        typer.Option(
            "--timeout",
            "-o",
            help="Maximum number of seconds before the process is killed.",
        ),
    ] = None,
    profile: PROFILE_OPTION = None,
    interactive: Annotated[
        bool,
        typer.Option("--interactive", "-i", is_flag=True, help="Interactive mode."),
    ] = False,
) -> None:
    """
    Adds a new command with an alias, a template, description, and tags. The command
    can be created in interactive if no options are provided or the `--interactive`
    flag is used.
    """
    log.debug("cmd.add called. alias=%s, interactive=%s", alias, interactive)
    add_cmd_args = command_handlers.AddCommandArgs(
        alias=alias,
        template=template,
        description=description,
        tags=tags,
        cwd=cwd,
        shell=shell,
        env=env,
        timeout=timeout,
        interactive=interactive,
        profile=profile,
    )
    command_handlers.run_add_command(
        args=add_cmd_args,
        get_cmd_services=container.get_command_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("get")
@cli_guard
def get(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to retrieve.",
            autocompletion=complete_command_aliases,
        ),
    ],
    profile: PROFILE_OPTION = None,
) -> None:
    """
    Retrieves and displays a saved command stored under the provided alias.
    """
    log.debug("cmd.get called. alias=%s", alias)
    command_handlers.run_get_command(
        alias=alias,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_console=container.get_console,
    )


@app.command("update")
@cli_guard
def update(
    current_alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to update.",
            autocompletion=complete_command_aliases,
        ),
    ],
    template: Annotated[
        str | None, typer.Option("--template", "-t", help="The new template.")
    ] = None,
    description: Annotated[
        str | None, typer.Option("--description", "-d", help="The new description.")
    ] = None,
    new_alias: Annotated[
        str | None, typer.Option("--alias", "-a", help="The new alias.")
    ] = None,
    cwd: Annotated[
        str | None,
        typer.Option("--cwd", help="Working directory to run the command from."),
    ] = None,
    clear_cwd: Annotated[
        bool, typer.Option("--clear-cwd", help="Clear the stored working directory.")
    ] = False,
    shell: Annotated[
        str | None,
        typer.Option("--shell", help="Shell to use when running the command."),
    ] = None,
    clear_shell: Annotated[
        bool, typer.Option("--clear-shell", help="Clear the stored shell.")
    ] = False,
    env: Annotated[
        list[str] | None,
        typer.Option(
            "--env",
            help="Environment variable to set when running the command, in KEY=VALUE format.",
        ),
    ] = None,
    clear_env: Annotated[
        bool, typer.Option("--clear-env", help="Clear stored environment variables.")
    ] = False,
    timeout: Annotated[
        int | None,
        typer.Option(
            "--timeout",
            help="Maximum number of seconds before the process is killed.",
        ),
    ] = None,
    clear_timeout: Annotated[
        bool, typer.Option("--clear-timeout", help="Clear the stored timeout.")
    ] = False,
    set_: Annotated[
        list[str] | None,
        typer.Option(
            "--set",
            "-s",
            help="A list of key=value pairs to update.",
            autocompletion=command_editable_field_options,
        ),
    ] = None,
    edit_mode: Annotated[
        bool,
        typer.Option("--edit", "-e", help="Edit mode."),
    ] = False,
    edit_fields: Annotated[
        str | None,
        typer.Option(
            "--edit-fields",
            help="A list of fields to be edited in edit mode, separated by commas. Defaults to all fields.",
        ),
    ] = None,
) -> None:
    """
    Updates an existing command with the provided options. Each field can be updated
    individually or in bulk using the `--set` option. Using the `--edit` option enables
    editing the already stored primary values (alias, template, and description) in an
    interactive mode.
    """
    log.debug(
        "cmd.update called. alias=%s template_provided=%s description_provided=%s new_alias_provided=%s "
        "set_used=%s, edit_mode=%s, edit_fields=%s",
        current_alias,
        template is not None,
        description is not None,
        new_alias is not None,
        set_ is not None,
        edit_mode,
        edit_fields,
    )
    command_handlers.run_update_command(
        current_alias=current_alias,
        template=template,
        description=description,
        new_alias=new_alias,
        cwd=cwd,
        clear_cwd=clear_cwd,
        shell=shell,
        clear_shell=clear_shell,
        env=env,
        clear_env=clear_env,
        timeout=timeout,
        clear_timeout=clear_timeout,
        set_pairs=set_,
        edit_mode=edit_mode,
        edit_fields=edit_fields,
        get_cmd_services=container.get_command_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
    )


app.command("edit", hidden=True)(update)


@app.command("list")
@cli_guard
def list_cmds(
    order: Annotated[
        Optional[str],
        typer.Option(
            "--order",
            "-o",
            help="The field to order the results by.",
            autocompletion=command_field_options,
        ),
    ] = None,
    tags: Annotated[
        list[str] | None,
        typer.Option(
            "--tag",
            "-t",
            help="The tag to filter by.",
            autocompletion=complete_tag_names,
        ),
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
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults fields in settings.",
            autocompletion=command_field_options,
        ),
    ] = None,
    profile: PROFILE_OPTION = None,
) -> None:
    """
    Displays all stored commands in a list format. The number of results can be limited
    with the `--limit` option. The output fields can be customized with the `--field` option.
    """
    log.debug(
        "cmd.list called. order=%s, tags=%s, limit=%s, fields=%s, page=%s",
        order,
        tags,
        limit,
        fields,
        page,
    )
    command_handlers.run_list_command(
        limit=limit,
        order=order,
        tags=tags,
        fields=fields,
        page=page,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_command_display_field_resolver,
    )


app.command("ls", hidden=True)(list_cmds)


@app.command("search")
@cli_guard
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        Optional[int], typer.Option(help="The maximum number of results to return.")
    ] = None,
    page: Annotated[
        bool | None,
        typer.Option(
            "--page/--no-page",
            help="Force or disable the interactive pager, overriding the pager mode in settings.",
        ),
    ] = None,
    search_fields: Annotated[
        list[str],
        typer.Option(
            "--in",
            "-i",
            help="The fields to search within.",
            autocompletion=command_field_options,
        ),
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults to all fields.",
            autocompletion=command_field_options,
        ),
    ] = None,
    profile: PROFILE_OPTION = None,
) -> None:
    """
    Searches the database for commands matching the provided search term. The search fields
    can be customized with the `--in` option. The output fields can be customized with the
    `--field` option.
    """
    log.debug(
        "cmd.search called. term=%s, limit=%s, search_fields=%s, fields=%s, page=%s",
        term,
        limit,
        search_fields,
        fields,
        page,
    )
    command_handlers.run_search_command(
        term=term,
        limit=limit,
        search_fields=search_fields,
        fields=fields,
        page=page,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_command_display_field_resolver,
        get_search_field_resolver=container.get_command_search_field_resolver,
    )


app.command("find", hidden=True)(search)


@app.command("delete")
@cli_guard
def delete(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to delete.",
            autocompletion=complete_command_aliases,
        ),
    ],
    profile: PROFILE_OPTION = None,
) -> None:
    """
    Deletes the command stored under the provided alias.
    """
    log.debug("cmd.delete called. alias=%s", alias)
    command_handlers.run_delete_command(
        alias=alias,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_console=container.get_console,
    )


app.command("del", hidden=True)(delete)
app.command("rm", hidden=True)(delete)
app.command("remove", hidden=True)(delete)


@app.command("tag")
@cli_guard
def add_tags(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to tag.",
            autocompletion=complete_command_aliases,
        ),
    ] = None,
    tags: Annotated[
        list[str],
        typer.Argument(
            help="The tags to add to the command, separated by commas.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
    profile: PROFILE_OPTION = None,
) -> None:
    """
    Adds the provided tags to the command stored under the provided alias. Tags must
    be existing.
    """
    log.debug("cmd.tag.add called. alias=%s, tags=%s", alias, tags)
    command_handlers.run_attach_tags(
        alias=alias,
        tag_names=tags,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("untag")
@cli_guard
def remove_tags(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to untag.",
            autocompletion=complete_command_aliases,
        ),
    ] = None,
    tags: Annotated[
        list[str],
        typer.Argument(
            help="The tags to remove from the command, separated by commas.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
    profile: PROFILE_OPTION = None,
) -> None:
    """
    Removes the provided tags from the command stored under the provided alias.
    """
    log.debug("cmd.tag.remove called. alias=%s, tags=%s", alias, tags)
    command_handlers.run_detach_tags(
        alias=alias,
        tag_names=tags,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("move")
@cli_guard
def move(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to move.",
            autocompletion=complete_command_aliases,
        ),
    ],
    target_profile: Annotated[
        str,
        typer.Argument(
            help="The profile to move the command to.",
            autocompletion=complete_profile_names,
        ),
    ],
    profile: PROFILE_OPTION = None,
) -> None:
    log.debug(
        "cmd.move called. alias=%s, target_profile=%s, profile=%s",
        alias,
        target_profile,
        profile,
    )
    command_handlers.run_move_command(
        alias=alias,
        target_profile=target_profile,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_console=container.get_console,
    )


@app.command("copy")
@cli_guard
def copy(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to copy.",
            autocompletion=complete_command_aliases,
        ),
    ],
    target_profile: Annotated[
        str,
        typer.Argument(
            help="The profile to copy the command into.",
            autocompletion=complete_profile_names,
        ),
    ],
    new_alias: Annotated[
        Optional[str],
        typer.Option(
            "--as",
            help="New alias for the copy, if it should different than the original.",
        ),
    ] = None,
    profile: PROFILE_OPTION = None,
) -> None:
    log.debug(
        "cmd.copy called. alias=%s, target_profile=%s, new_alias=%s, profile=%s",
        alias,
        target_profile,
        new_alias,
        profile,
    )
    command_handlers.run_copy_command(
        alias=alias,
        target_profile=target_profile,
        new_alias=new_alias,
        profile=profile,
        get_cmd_services=container.get_command_services,
        get_console=container.get_console,
    )
