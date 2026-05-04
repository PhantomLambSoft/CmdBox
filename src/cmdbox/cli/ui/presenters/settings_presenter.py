from dataclasses import asdict

from cmdbox.cli.ui.primitives import stack, kv_table, section
from cmdbox.settings.models import Settings


def render_settings_show(settings: Settings, fields: list[str] | None = None):
    """
    Renders the settings to display either fully or partially, based on the input.

    This function takes in a settings object and an optional list of fields. If the
    fields list is provided, only the specified fields from the settings object
    will be rendered. Otherwise, the entire settings object will be rendered.

    Args:
        settings (Settings): The settings object to be rendered.
        fields (list[str] | None): An optional list of field names to render. If
            None, the full settings object is rendered.

    Returns:
        str: The rendered representation of the settings.
    """
    if fields is None:
        return render_full_settings(settings)
    else:
        return render_settings_fields(settings, fields)


def render_full_settings(settings: Settings):
    """
    Renders the full configuration settings into a styled section format.

    This function iterates through the provided settings object, extracts sections
    and their respective values, and organizes them into a structured visual
    presentation. Each section is assigned a styled border to enhance readability.

    Args:
        settings (Settings): The settings object containing configuration data,
            structured as key-value pairs.

    Returns:
        Section: A styled section containing all rendered sub-sections representing
            the settings in a visually organized format.
    """
    ret = []
    for _section, values in asdict(settings).items():
        ret.append(
            section(
                title=_section, body=render_section(values), border_style="status.info"
            )
        )
    return section(title="Settings", body=stack(*ret), border_style="status.info")


def render_section(_section: dict):
    """
    Renders a nested dictionary structure into a visual representation using specific rendering functions.

    This function processes a nested dictionary and generates a rendered output. It recursively handles
    dictionaries as nested sections and renders key-value pairs as a table. Special handling is provided
    for dictionary items with the key 'alias_mapping'.

    Args:
        _section (dict): The nested dictionary structure to be rendered.

    Returns:
        Any: The final rendered output based on the processed structure.
    """
    ret = []
    for key, value in _section.items():
        if isinstance(value, dict):
            sub = render_section(value)
            if key == "alias_mapping":
                body = sub
            else:
                body = section(title=key, body=sub)
        else:
            body = kv_table(rows=[(key, value)])
        ret.append(body)
    return stack(*ret)


def render_settings_fields(settings: Settings, fields: list[str]):
    """
    Renders the settings fields as a formatted table.

    This function takes settings and a list of field names, extracts the relevant
    fields from the settings object, and displays them as rows in a table. The
    output is a styled section containing the table.

    Args:
        settings (Settings): An object containing configuration settings, structured
            as fields and their respective values.
        fields (list[str]): A list of field names to extract and display from the
            settings object.

    Returns:
        str: A styled section containing a table representation of the settings
        fields.
    """
    display_fields = []
    for key, value in asdict(settings).items():
        display_fields.extend(parse_settings_fields(value, fields))
    table = kv_table(rows=display_fields)
    return section(title="Settings", body=table, border_style="status.info")


def parse_settings_fields(_section: dict, fields: list[str]):
    """
    Parses a nested dictionary to extract specific fields and their
    associated values.

    This function traverses a given dictionary recursively and collects
    key-value pairs from the dictionary that match the specified fields.

    Args:
        _section (dict): A nested dictionary to search for matching fields.
        fields (list[str]): A list of field names to look for in _section.

    Returns:
        list[tuple]: A list of tuples containing key-value pairs for
        the fields found in _section, including from nested dictionaries.
    """
    ret = []
    for key, value in _section.items():
        if isinstance(value, dict):
            ret.extend(parse_settings_fields(value, fields))
        if key in fields:
            ret.append((key, value))
    return ret
