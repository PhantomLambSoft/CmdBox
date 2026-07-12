import logging
from typing import Annotated, Optional

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.completions.fields import variable_field_options
from cmdbox.cli.completions.variables import complete_variable_names
from cmdbox.cli.completions.tags import complete_tag_names
from cmdbox.cli.handlers import variable_handlers

app = typer.Typer(no_args_is_help=True)

cli_guard = make_cli_guard(container.get_console)

log = logging.getLogger(__name__)


@app.command("add")
@cli_guard
def add(
    name: Annotated[str, typer.Argument(help="The name of the variable.")] = None,
    value: Annotated[str, typer.Argument(help="The value of the variable.")] = None,
    tags: Annotated[
        list[str],
        typer.Option(
            "--tags",
            "-t",
            help="A list of tags to associate with the variable, separated by commas.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
    interactive: Annotated[
        bool,
        typer.Option("--interactive", "-i", is_flag=True, help="Interactive mode."),
    ] = False,
) -> None:
    """
    Adds a new variable with the specified name, value, and tags. The variable
    can be created in interactive mode if no options are provided or the `--interactive`
    flag is used.
    """
    log.debug(
        "var.add called. name=%s, value=%s, tags=%s, interactive=%s",
        name,
        value,
        tags,
        interactive,
    )
    add_var_args = variable_handlers.AddVariableArgs(
        name=name, value=value, tags=tags, interactive=interactive
    )
    variable_handlers.run_add_variable(
        args=add_var_args,
        get_var_services=container.get_variable_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("get")
@cli_guard
def get(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the variable to retrieve.",
            autocompletion=complete_variable_names,
        ),
    ],
) -> None:
    """
    Retrieves and displays the variable stored under the provided name.
    """
    log.debug("var.get called. name=%s", name)
    variable_handlers.run_get_variable(
        name=name,
        get_var_services=container.get_variable_services,
        get_console=container.get_console,
    )


@app.command("update")
@cli_guard
def update(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the variable to update.",
            autocompletion=complete_variable_names,
        ),
    ],
    value: Annotated[str, typer.Option("--value", "-v", help="The new value.")] = None,
    new_name: Annotated[str, typer.Option("--name", "-n", help="The new name.")] = None,
    set_: Annotated[
        list[str],
        typer.Option(
            "--set",
            "-s",
            help="A list of key=value pairs to update.",
            autocompletion=variable_field_options,
        ),
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
    Updates an existing variable with the provided options. Each field can be updated
    individually or in bulk using the `--set` option. Using the `--edit` option enables
    editing the already stored values in an interactive mode.
    """
    log.debug(
        "var.update called. name=%s, value=%s, new_name=%s, set_pairs=%s, edit_mode=%s, edit_fields=%s",
        name,
        value,
        new_name,
        set_,
        edit_mode,
        edit_fields,
    )
    variable_handlers.run_update_variable(
        name=name,
        value=value,
        new_name=new_name,
        set_pairs=set_,
        edit_mode=edit_mode,
        edit_fields=edit_fields,
        get_var_services=container.get_variable_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
    )


app.command("edit", hidden=True)(update)


@app.command("list")
@cli_guard
def list_vars(
    order: Annotated[
        str,
        typer.Option(
            "--order",
            "-o",
            help="The field to order the results by.",
            autocompletion=variable_field_options,
        ),
    ] = "name",
    tags: Annotated[
        list[str],
        typer.Option(
            "--tag",
            "-t",
            help="The tags to filter by.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
    limit: Annotated[
        Optional[int],
        typer.Option("--limit", "-l", help="The maximum number of results to return."),
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The fields to display in the results list.",
            autocompletion=variable_field_options,
        ),
    ] = None,
) -> None:
    """
    Displays all stored variables in a list format. The number of results can be limited
    with the `--limit` option. The output fields can be customized with the `--field` option.
    """
    log.debug(
        "var.list called. order=%s, tags=%s, limit=%s, fields=%s",
        order,
        tags,
        limit,
        fields,
    )
    variable_handlers.run_list_variables(
        order_by=order,
        tags=tags,
        limit=limit,
        fields=fields,
        get_var_services=container.get_variable_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_variable_display_field_resolver,
    )


app.command("ls", hidden=True)(list_vars)


@app.command("search")
@cli_guard
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        Optional[int], typer.Option(help="The maximum number of results to return.")
    ] = None,
    search_fields: Annotated[
        list[str],
        typer.Option(
            "--in",
            "-i",
            help="The fields to search within.",
            autocompletion=variable_field_options,
        ),
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults to all fields.",
            autocompletion=variable_field_options,
        ),
    ] = None,
) -> None:
    """
    Searches the database for variables matching the provided search term. The search fields
    can be customized with the `--in` option. The output fields can be customized with the
    `--field` option.
    """
    log.debug(
        "var.search called. term=%s, limit=%s, search_fields=%s, fields=%s",
        term,
        limit,
        search_fields,
        fields,
    )
    variable_handlers.run_search_variables(
        term=term,
        limit=limit,
        search_fields=search_fields,
        fields=fields,
        get_var_services=container.get_variable_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_variable_display_field_resolver,
        get_search_field_resolver=container.get_variable_search_field_resolver,
    )


app.command("find", hidden=True)(search)


@app.command("delete")
@cli_guard
def delete(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the variable to delete.",
            autocompletion=complete_variable_names,
        ),
    ],
) -> None:
    """
    Deletes the variable stored under the provided name.
    """
    log.debug("var.delete called. name=%s", name)
    variable_handlers.run_delete_variable(
        name=name,
        get_var_services=container.get_variable_services,
        get_console=container.get_console,
    )


app.command("del", hidden=True)(delete)
app.command("rm", hidden=True)(delete)
app.command("remove", hidden=True)(delete)


@app.command("tag")
@cli_guard
def add_tags(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the command to tag.",
            autocompletion=complete_variable_names,
        ),
    ],
    tags: Annotated[
        list[str],
        typer.Argument(
            help="The tags to add to the command.", autocompletion=complete_tag_names
        ),
    ] = None,
) -> None:
    """
    Adds the provided tags to the command stored under the provided alias. Tags must
    be existing.
    """
    log.debug("var.tag.add called. name=%s, tags=%s", name, tags)
    variable_handlers.run_attach_tags(
        name=name,
        tag_names=tags,
        get_var_services=container.get_variable_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("untag")
@cli_guard
def remove_tags(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the command to untag.",
            autocompletion=complete_variable_names,
        ),
    ],
    tags: Annotated[
        list[str],
        typer.Argument(
            help="The tags to remove from the command.",
            autocompletion=complete_tag_names,
        ),
    ] = None,
) -> None:
    """
    Removes the provided tags from the command stored under the provided alias.
    """
    log.debug("var.tag.remove called. name=%s, tags=%s", name, tags)
    variable_handlers.run_detach_tags(
        name=name,
        tag_names=tags,
        get_var_services=container.get_variable_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )
