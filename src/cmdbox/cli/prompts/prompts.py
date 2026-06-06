import sys
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
        "Enter template: ",
        prompt_continuation="            ... ",
        multiline=True,
        validator=validator,
        default=default,
    )
    return template


def prompt_for_description(default: str = "") -> str:
    description = prompt("Enter description: ", default=default)
    return description


def prompt_for_cwd(default: str = "") -> str:
    cwd = prompt("Enter working directory: ", default=default)
    return cwd


def prompt_for_shell(default: str = "") -> str:
    shell = prompt("Enter shell: ", default=default)
    return shell


def prompt_for_timeout(default: str = "") -> str:
    timeout = prompt("Enter timeout (in seconds): ", default=default)
    return timeout


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


def prompt_for_missing_var(var_name: str) -> str:
    """
    Prompts the user to input a value for a given variable name.

    This function checks if the standard output is a terminal. If it is, the function
    uses the `prompt()` mechanism to get the input. If standard output is not a terminal
    (e.g., the command is being run in cbe emit mode), the function writes the prompt to
    standard error, flushes the stream, and then reads input directly from the console.

    Args:
        var_name: A string representing the name of the variable for which a value
            is being requested.

    Returns:
        A string containing the value entered by the user for the specified variable.
    """
    message = f"Enter value for <{var_name}>: "

    # Normal execution
    if sys.stdout.isatty():
        return prompt(message)

    # Stdout is a pipe, not tty (e.g. cbe emit mode) - open console directly
    sys.stderr.write(message)
    sys.stderr.flush()
    return sys.stdin.readline().strip("\n")
