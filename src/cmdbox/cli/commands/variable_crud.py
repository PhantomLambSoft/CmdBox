from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.prompts.completers import TagCompleter
from cmdbox.cli.prompts.prompts import (
    prompt_for_name,
    prompt_for_value,
    prompt_for_tags,
)
from cmdbox.cli.prompts.validators import TagNameValidator
from cmdbox.repositories.errors import UnknownNameError

app = typer.Typer(no_args_is_help=True)


@app.command("add")
def add(
    name: Annotated[str, typer.Argument(help="The name of the variable.")] = None,
    value: Annotated[str, typer.Argument(help="The value of the variable.")] = None,
    tags: Annotated[
        list[str],
        typer.Option(
            "--tags",
            "-t",
            help="A list of tags to associate with the variable, separated by commas.",
        ),
    ] = None,
    interactive: Annotated[
        bool,
        typer.Option("--interactive", "-i", is_flag=True, help="Interactive mode."),
    ] = False,
):
    if interactive or name is None:
        name = prompt_for_name()
    if interactive or value is None:
        value = prompt_for_value()
    if interactive or tags is None:
        tag_services = container.get_tag_services()

        def get_tags(query: str) -> list[str]:
            found_tags = tag_services.search(query, fields="name")
            return [tag.name for tag in found_tags]

        tag_completer = TagCompleter(get_tags)
        tag_validator = TagNameValidator()
        tags = prompt_for_tags(tag_completer, tag_validator)
    if not tags:
        tags = None
    var_service = container.get_variable_services()
    var = var_service.create_variable(
        name=name,
        value=value,
        tags=tags,
    )
    console = container.get_console()
    console.success("Variable successfully created:")
    console.print_variable(var)


@app.command("get")
def get(
    name: Annotated[str, typer.Argument(help="The name of the variable to retrieve.")],
):
    console = container.get_console()
    var_service = container.get_variable_services()
    try:
        var = var_service.get_variable(name)
        console.print_variable(var)
    except UnknownNameError:
        console.error(f"Variable '{name}' not found.")


@app.command("update")
def update():
    console = container.get_console()
    console.error("Not yet implemented.")


@app.command("list")
def list_vars(
    order: Annotated[
        str, typer.Option("--order", "-o", help="The field to order the results by.")
    ] = "name",
    tags: Annotated[
        list[str], typer.Option("--tag", "-t", help="The tags to filter by.")
    ] = None,
    limit: Annotated[
        int,
        typer.Option("--limit", "-l", help="The maximum number of results to return."),
    ] = 10,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field", "-f", help="The fields to display in the results list."
        ),
    ] = None,
):
    console = container.get_console()
    var_service = container.get_variable_services()
    var_list = var_service.list_variables(order_by=order, tags=tags, limit=limit)
    console.print_variable_list(var_list, output_fields=fields)


@app.command("search")
def search(
    term: Annotated[str, typer.Argument(help="The search term to use.")],
    limit: Annotated[
        int, typer.Option(help="The maximum number of results to return.")
    ] = 10,
    search_fields: Annotated[
        list[str], typer.Option("--in", "-i", help="The fields to search within.")
    ] = None,
    fields: Annotated[
        list[str] | None,
        typer.Option(
            "--field",
            "-f",
            help="The field(s) to display in the results list. Defaults to all fields.",
        ),
    ] = None,
):
    console = container.get_console()
    var_service = container.get_variable_services()
    var_list = var_service.search(term, fields=search_fields, limit=limit)
    console.print_variable_list(var_list, output_fields=fields)


@app.command("delete")
def delete(
    name: Annotated[str, typer.Argument(help="The name of the variable to delete.")],
):
    console = container.get_console()
    var_service = container.get_variable_services()
    var = var_service.get_variable(name)
    if var_service.delete_variable(name):
        console.success("Variable deleted successfully.")
        console.print_variable(var)
    else:
        console.error(f"Variable '{name}' not found.")
