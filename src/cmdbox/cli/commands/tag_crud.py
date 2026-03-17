import logging
from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.completions.fields import tag_editable_field_options, tag_field_options
from cmdbox.cli.completions.tags import complete_tag_names
from cmdbox.cli.handlers import tag_handlers

app = typer.Typer(no_args_is_help=True)

cli_guard = make_cli_guard(container.get_console)

log = logging.getLogger(__name__)


@app.command("add")
@cli_guard
def add(
    name: Annotated[
        str,
        typer.Argument(help="The name of the tag.", autocompletion=complete_tag_names),
    ] = None,
    description: Annotated[
        str, typer.Argument(help="A description of the tag.")
    ] = None,
    interactive: Annotated[
        bool,
        typer.Option(
            "--interactive",
            "-i",
            is_flag=True,
            help="Prompt for tag details interactively.",
        ),
    ] = False,
) -> None:
    """
    Adds a new tag with a name and description. The tag can be created in
    interactive if no options are provided or the `--interactive` flag is
    used.
    """
    log.debug(
        "tag.add called. name=%s, description=%s, interactive=%s",
        name,
        description,
        interactive,
    )
    add_tag_args = tag_handlers.AddTagArgs(
        name=name, description=description, interactive=interactive
    )
    tag_handlers.run_add_tag(
        args=add_tag_args,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("get")
@cli_guard
def get(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the tag to retrieve.", autocompletion=complete_tag_names
        ),
    ],
) -> None:
    """
    Gets and displays a saved tag under the provided name.
    """
    log.debug("tag.get called. name=%s", name)
    tag_handlers.run_get_tag(
        name=name,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


@app.command("update")
@cli_guard
def update(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the tag to update.", autocompletion=complete_tag_names
        ),
    ],
    description: Annotated[
        str,
        typer.Option("--description", "-d", help="The new description of the tag."),
    ] = None,
    new_name: Annotated[
        str,
        typer.Option("--name", "-n", help="The new name of the tag."),
    ] = None,
    set_: Annotated[
        list[str],
        typer.Option(
            "--set",
            "-s",
            help="A list of key=value pairs to update.",
            autocompletion=tag_editable_field_options,
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
            "-ef",
            help="A list of fields to be edited in edit mode, separated by commas. Defaults to all fields.",
        ),
    ] = None,
) -> None:
    """
    Updates an existing tag with the provided options. Each field can be updated
    individually or in bulk using the `--set` option. Using the `--edit` option enables
    editing the already stored values in an interactive mode.
    """
    log.debug(
        "tag.update called. name=%s, description=%s, new_name=%s, set_pairs=%s, edit_mode=%s, edit_fields=%s",
        name,
        description,
        new_name,
        set_,
        edit_mode,
        edit_fields,
    )
    tag_handlers.run_update_tag(
        name=name,
        description=description,
        new_name=new_name,
        set_pairs=set_,
        edit_mode=edit_mode,
        edit_fields=edit_fields,
        get_tag_services=container.get_tag_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
    )


app.command("edit", hidden=True)(update)


@app.command("list")
@cli_guard
def list_tags(
    order: Annotated[
        str,
        typer.Option(
            "--order",
            "-o",
            help="The field to order the results by.",
            autocompletion=tag_field_options,
        ),
    ] = "name",
    limit: Annotated[
        int,
        typer.Option("--limit", "-l", help="The maximum number of results to return."),
    ] = 10,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The fields to display in the results list.",
            autocompletion=tag_field_options,
        ),
    ] = None,
) -> None:
    """
    Displays all stored tags in a list format. The number of results can be limited
    with the `--limit` option. The output fields can be customized with the `--field` option.
    """
    log.debug("tag.list called. order=%s, limit=%s, fields=%s", order, limit, fields)
    tag_handlers.run_list_tags(
        limit=limit,
        fields=fields,
        order_by=order,
        get_tag_services=container.get_tag_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_tag_display_field_resolver,
    )


app.command("ls", hidden=True)(list_tags)


@app.command("search")
@cli_guard
def search(
    term: Annotated[
        str,
        typer.Argument(
            help="The search term to use.", autocompletion=complete_tag_names
        ),
    ],
    limit: Annotated[
        int, typer.Option(help="The maximum number of results to return.")
    ] = 10,
    search_fields: Annotated[
        list[str],
        typer.Option(
            "--in",
            "-i",
            help="The fields to search within.",
            autocompletion=tag_field_options,
        ),
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults to all fields.",
            autocompletion=tag_field_options,
        ),
    ] = None,
) -> None:
    """
    Searches the database for tags matching the provided search term. The search fields
    can be customized with the `--in` option. The output fields can be customized with the
    `--field` option.
    """
    log.debug(
        "tag.search called. term=%s, limit=%s, search_fields=%s, fields=%s",
        term,
        limit,
        search_fields,
        fields,
    )
    tag_handlers.run_search_tags(
        term=term,
        limit=limit,
        search_fields=search_fields,
        fields=fields,
        get_tag_services=container.get_tag_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
        get_display_field_resolver=container.get_tag_display_field_resolver,
        get_search_field_resolver=container.get_tag_search_field_resolver,
    )


app.command("find", hidden=True)(search)


@app.command("delete")
@cli_guard
def delete(
    name: Annotated[
        str,
        typer.Argument(
            help="The name of the tag to delete.", autocompletion=complete_tag_names
        ),
    ],
) -> None:
    """
    Deletes the tag stored under the provided name.
    """
    log.debug("tag.delete called. name=%s", name)
    tag_handlers.run_delete_tag(
        name=name,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )


app.command("del", hidden=True)(delete)
app.command("rm", hidden=True)(delete)
app.command("remove", hidden=True)(delete)
