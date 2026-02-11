from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.completions.fields import variable_field_options
from cmdbox.cli.completions.variables import complete_variable_names
from cmdbox.cli.completions.tags import complete_tag_names
from cmdbox.cli.handlers import variable_handlers


app = typer.Typer(no_args_is_help=True)

cli_guard = make_cli_guard(container.get_console)


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
) -> None:
    variable_handlers.run_update_variable(
        name=name,
        value=value,
        new_name=new_name,
        set_pairs=set_,
        get_var_services=container.get_variable_services,
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
        int,
        typer.Option("--limit", "-l", help="The maximum number of results to return."),
    ] = 10,
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
    variable_handlers.run_list_variables(
        order_by=order,
        tags=tags,
        limit=limit,
        fields=fields,
        get_var_services=container.get_variable_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
    )


app.command("ls", hidden=True)(list_vars)


@app.command("search")
@cli_guard
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        int, typer.Option(help="The maximum number of results to return.")
    ] = 10,
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
    variable_handlers.run_search_variables(
        term=term,
        limit=limit,
        search_fields=search_fields,
        fields=fields,
        get_var_services=container.get_variable_services,
        get_settings=container.get_settings,
        get_console=container.get_console,
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
    variable_handlers.run_detach_tags(
        name=name,
        tag_names=tags,
        get_var_services=container.get_variable_services,
        get_tag_services=container.get_tag_services,
        get_console=container.get_console,
    )
