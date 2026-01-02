from prompt_toolkit import prompt

from cmdbox.cli.prompts.completers import TagCompleter
from cmdbox.cli.prompts.validators import (
    AliasValidator,
    TemplateValidator,
    TagNameValidator,
    NameValidator,
)


def prompt_for_alias(validator: AliasValidator) -> str:
    alias = prompt("Enter alias: ", validator=validator)
    return alias


def prompt_for_template(validator: TemplateValidator) -> str:
    template = prompt("Enter template: ", multiline=True, validator=validator)
    return template


def prompt_for_description() -> str:
    description = prompt("Enter description: ")
    return description


def prompt_for_name(validator: NameValidator | TagNameValidator) -> str:
    name = prompt("Enter name: ", validator=validator)
    return name


def prompt_for_value() -> str:
    value = prompt("Enter variable value: ")
    return value


def prompt_for_tags(
    tag_completer: TagCompleter, validator: TagNameValidator
) -> list[str] | None:
    tags = prompt(
        "Enter tags (comma-separated): ", completer=tag_completer, validator=validator
    )
    if not tags:
        return None
    return tags.split(",")
