from prompt_toolkit import prompt

from cmdbox.cli.prompts.completers import TagCompleter
from cmdbox.cli.prompts.validators import (
    AliasValidator,
    TemplateValidator,
    TagNameValidator,
    NameValidator,
)


def prompt_for_confirm(message: str, default: bool = False) -> bool:
    suffix = " [Y/n]: " if default else " [y/N]: "
    result = prompt(message + suffix).strip().lower()
    if result == "":
        return default
    return result in ["y", "yes"]


def prompt_for_alias(validator: AliasValidator, default: str = "") -> str:
    alias = prompt("Enter alias: ", validator=validator, default=default)
    return alias


def prompt_for_template(validator: TemplateValidator, default: str = "") -> str:
    template = prompt(
        "Enter template: ", multiline=True, validator=validator, default=default
    )
    return template


def prompt_for_description(default: str = "") -> str:
    description = prompt("Enter description: ", default=default)
    return description


def prompt_for_name(
    validator: NameValidator | TagNameValidator, default: str = ""
) -> str:
    name = prompt("Enter name: ", validator=validator, default=default)
    return name


def prompt_for_value(default: str = "") -> str:
    value = prompt("Enter variable value: ", default=default)
    return value


def prompt_for_tags(
    tag_completer: TagCompleter, validator: TagNameValidator, default: str = ""
) -> list[str] | None:
    tags = prompt(
        "Enter tags (comma-separated): ",
        completer=tag_completer,
        validator=validator,
        default=default,
    )
    if not tags:
        return None
    return tags.split(",")
