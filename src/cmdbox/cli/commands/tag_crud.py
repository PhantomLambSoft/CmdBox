from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.prompts.prompts import prompt_for_name, prompt_for_description
from cmdbox.cli.prompts.validators import TagNameValidator
from cmdbox.repositories.errors import UnknownNameError

app = typer.Typer(no_args_is_help=True)


@app.command("add")
def add(
    name: Annotated[str, typer.Argument(help="The name of the tag.")] = None,
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
):
    if interactive or not name:
        validator = TagNameValidator()
        name = prompt_for_name(validator)
    if interactive or not description:
        description = prompt_for_description()
    tag_service = container.get_tag_services()
    tag = tag_service.create_tag(name=name, description=description)
    console = container.get_console()
    console.success("Tag successfully created:")
    console.print_tag(tag)


@app.command("get")
def get(
    name: Annotated[str, typer.Argument(help="The name of the tag to retrieve.")],
):
    console = container.get_console()
    tag_service = container.get_tag_services()
    try:
        tag = tag_service.get_tag(name)
        console.print_tag(tag)
    except UnknownNameError:
        console.error(f"Tag '{name}' not found.")
        return


@app.command("update")
def update():
    console = container.get_console()
    console.error("Not yet implemented.")


@app.command("list")
def list_tags(
    order: Annotated[
        str, typer.Option("--order", "-o", help="The field to order the results by.")
    ] = "name",
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
    tag_service = container.get_tag_services()
    tag_list = tag_service.list_tags(order_by=order, limit=limit)
    console.print_tag_list(tag_list, output_fields=fields)


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
    tag_service = container.get_tag_services()
    tag_list = tag_service.search(term, fields=search_fields, limit=limit)
    console.print_tag_list(tag_list, output_fields=fields)


@app.command("delete")
def delete(
    name: Annotated[str, typer.Argument(help="The name of the tag to delete.")],
):
    console = container.get_console()
    tag_service = container.get_tag_services()
    tag = tag_service.get_tag(name)
    if tag_service.delete_tag(name):
        console.success("Tag deleted successfully.")
        console.print_tag(tag)
    else:
        console.error(f"Tag '{name}' not found.")
