from dataclasses import dataclass
from datetime import date, datetime
from typing import Any, Mapping, Sequence

from rich import box
from rich.align import Align
from rich.console import Group, RenderableType
from rich.padding import Padding
from rich.panel import Panel
from rich.rule import Rule
from rich.table import Table
from rich.text import Text


@dataclass(frozen=True)
class UiStyles:
    # General
    title: str = "ui.title"
    subtitle: str = "ui.subtitle"
    muted: str = "ui.muted"
    dim: str = "ui.dim"

    # Status
    success: str = "status.success"
    info: str = "status.info"
    warning: str = "status.warning"
    error: str = "status.error"

    # Panels / borders
    panel_border: str = "ui.border"
    panel_title: str = "ui.panel_title"

    # Tables
    table_header: str = "ui.table_header"
    table_caption: str = "ui.caption"
    kv_key: str = "ui.kv.key"
    kv_value: str = "ui.kv.value"

    # Code-ish output (semantic)
    code: str = "code"
    code_inline: str = "code.inline"
    code_block: str = "code.block"


def to_text(value: Any, *, style: str | None = None) -> Text:
    """
    Converts a given value into a formatted textual representation, optionally applying
    a text style.

    The function takes various types of input (e.g., None, datetime, date, or other types)
    and returns a corresponding textual representation as a `Text` object. If a style
    is specified, it will stylize the output text.

    Args:
        value: The input value to convert. Can be of type None, datetime, date,
            or any type that can be converted into a string representation.
        style: An optional string specifying the style to apply to the output text.

    Returns:
        Text: A formatted textual representation of the input value.
    """
    if value is None:
        t = Text("null")
    elif isinstance(value, Text):
        if style:
            value.stylize(style)
        return value
    elif isinstance(value, datetime):
        t = Text(value.isoformat(sep=" ", timespec="seconds"))
    elif isinstance(value, date):
        t = Text(value.isoformat())
    else:
        t = Text(str(value))

    if style:
        t.stylize(style)
    return t


def spacer(lines: int = 1) -> RenderableType:
    return Text("\n" * max(0, lines))


def divider(title: str | None = None, *, style: str = "ui.dim") -> Rule:
    if title:
        return Rule(Text(title), style=style)
    return Rule(style=style)


def col(name: str, *, style: str = "", **kwargs) -> tuple[str, dict[str, Any]]:
    """
    A helper function that generates a table column.

    This function facilitates the creation of a tuple that consists of a string name and a
    dictionary. The dictionary contains a "style" key with a specified value and may also include
    other key-value pairs provided as additional arguments.

    Args:
        name (str): The name or identifier for the column.
        style (str, optional): A string specifying the style attribute. Defaults to an empty string.
        **kwargs: Additional key-value pairs to include in the dictionary.

    Returns:
        tuple[str, dict[str, Any]]: A tuple where the first element is the provided name, and the
        second element is a dictionary containing the "style" attribute along with any additional
        key-value pairs.
    """
    data = {"style": style}
    data.update(kwargs)
    return name, data


def banner(
    title: str,
    subtitle: str | None = None,
    *,
    status: str | None = None,
    styles: UiStyles = UiStyles(),
) -> RenderableType:
    """
    A consistent header block.

    status: one of "success" | "info" | "warning" | "error" or a Theme style name
    """
    if status:
        # Allow semantic names or full style names
        status_style = getattr(styles, status, status)
        border_style = status_style
        title_style = status_style
    else:
        border_style = styles.panel_border
        title_style = styles.title

    title_text = Text(title, style=title_style)

    parts: list[RenderableType] = [title_text]

    if subtitle:
        parts.append(Text(subtitle, style=styles.subtitle))

    return Panel(
        Group(*parts),
        border_style=border_style,
        padding=(0, 1),
    )


def tag_pill(
    name: str,
    *,
    style: str = "tag.pill",
    padding: tuple[int, int] = (0, 1),
) -> RenderableType:
    text = Text(name, style=style)
    return Padding(text, padding)


def tag_block(
    tags: Sequence[str],
    *,
    style: str = "tag.pill",
    separator: str = " ",
) -> RenderableType:
    text = Text()

    for i, tag in enumerate(tags):
        if i > 0:
            text.append(separator)
        text.append(f" {tag} ", style=style)

    return text


def status_line(
    message: str,
    *,
    status: str = "info",
    styles: UiStyles = UiStyles(),
) -> Text:
    """
    One line status message with semantic styling.
    """
    status_style = getattr(styles, status, status)
    return Text(message, style=status_style)


def bullet_list(
    items: Sequence[Any],
    *,
    bullet: str = "•",
    item_style: str = "ui.kv.value",
    bullet_style: str = "ui.muted",
    styles: UiStyles = UiStyles(),
    empty_message: str | None = None,
) -> RenderableType:
    """
    Simple bullet list that is safe against markup.
    """
    if not items:
        if empty_message is None:
            return Text("")
        return Text(empty_message, style=styles.muted)

    lines: list[RenderableType] = []
    for item in items:
        b = Text(f"{bullet} ", style=bullet_style if bullet_style else styles.muted)
        v = to_text(item, style=item_style if item_style else styles.kv_value)
        lines.append(Text.assemble(b, v))

    return Group(*lines)


def kv_table(
    rows: Sequence[tuple[str, Any]],
    *,
    key_style: str = "ui.kv.key",
    value_style: str = "ui.kv.value",
    show_header: bool = False,
    styles: UiStyles = UiStyles(),
) -> Table:
    """
    Two-column key/value table for detail views.
    """
    table = Table(
        show_header=show_header,
        box=None,
        pad_edge=False,
        padding=(0, 1),
        collapse_padding=True,
    )

    if show_header:
        table.add_column("Field", style=styles.table_header)
        table.add_column("Value", style=styles.table_header)
    else:
        table.add_column(
            justify="right", style=key_style or styles.kv_key, no_wrap=True
        )
        table.add_column(style=value_style or styles.kv_value, overflow="fold")

    for k, v in rows:
        table.add_row(
            to_text(k, style=key_style or styles.kv_key),
            to_text(v, style=value_style or styles.kv_value),
        )

    return table


def kv_panel(
    title: str,
    rows: Sequence[tuple[str, Any]],
    *,
    border_style: str = "ui.border",
    styles: UiStyles = UiStyles(),
) -> Panel:
    """
    Panel containing a key/value table.
    """
    return Panel(
        kv_table(rows, styles=styles),
        title=Text(title, style=styles.panel_title),
        border_style=border_style,
        padding=(0, 1),
    )


def table_panel(
    title: str,
    columns: Sequence[tuple[str, dict[str, Any]]],
    rows: Sequence[Sequence[Any]],
    *,
    caption: str | None = None,
    border_style: str = "ui.border",
    styles: UiStyles = UiStyles(),
) -> Panel:
    """
    Build a standard table inside a panel.

    columns: [("Alias", {"style": "...", "no_wrap": True}), ...]
    rows: list of row sequences. Each cell is converted with to_text.
    """
    table = Table(
        show_header=True,
        header_style=styles.table_header,
        pad_edge=False,
        collapse_padding=True,
        box=box.MINIMAL,
    )

    for name, kwargs in columns:
        table.add_column(name, **kwargs)

    for row in rows:
        table.add_row(*[to_text(cell) for cell in row])

    if caption:
        table.caption = caption
        table.caption_style = styles.table_caption

    return Panel(
        table,
        title=Text(title, style=styles.panel_title),
        border_style=border_style,
        padding=(0, 1),
    )


def code_inline(
    text: Any,
    *,
    style: str = "code.inline",
    styles: UiStyles = UiStyles(),
) -> Text:
    """
    Semantic inline code styling. Does not change terminal font, but gives a consistent look.
    """
    # If you do not define code.inline, fall back to code, then to a generic style.
    resolved = style
    if style == "code.inline":
        resolved = styles.code_inline or styles.code or "bold"
    return to_text(text, style=resolved)


def code_block(
    text: Any,
    *,
    title: str | None = None,
    border_style: str = "ui.border",
    style: str = "code.block",
    styles: UiStyles = UiStyles(),
    fit: bool = False,
) -> Panel:
    """
    Panel for code-ish blocks (templates, resolved commands, previews).
    """
    resolved = style
    if style == "code.block":
        resolved = styles.code_block or styles.code or "bold"

    body = to_text(text, style=resolved)

    panel_title = Text(title, style=styles.panel_title) if title else None
    panel_cls = Panel.fit if fit else Panel

    return panel_cls(
        body,
        title=panel_title,
        border_style=border_style,
        padding=(1, 1),
    )


def centered(renderable: RenderableType) -> RenderableType:
    return Align.center(renderable)


def stack(*renderables: RenderableType) -> RenderableType:
    """
    Group renderables vertically.
    """
    return Group(*renderables)


def section(
    title: str,
    body: RenderableType,
    caption: str | None = None,
    *,
    border_style: str = "ui.border",
    styles: UiStyles = UiStyles(),
) -> Panel:
    """
    Generic titled panel wrapper.
    """
    return Panel(
        body,
        title=Text(title, style=styles.panel_title),
        subtitle=Text(caption, style="ui.caption") if caption else None,
        border_style=border_style,
        padding=(0, 1),
    )


def summary_counts(
    counts: Mapping[str, int],
    *,
    styles: UiStyles = UiStyles(),
) -> RenderableType:
    """
    Small helper for displaying a few counts, like:
    Commands: 12  Variables: 4  Tags: 9
    """
    parts: list[Text] = []
    first = True
    for label, count in counts.items():
        if not first:
            parts.append(Text("  ", style=styles.muted))
        first = False
        parts.append(Text(f"{label}:", style=styles.muted))
        parts.append(Text(" ", style=styles.muted))
        parts.append(Text(str(count), style=styles.title))

    return Text.assemble(*parts)


def pluralize(count: int, singular: str, plural: str | None = None) -> str:
    if count == 1:
        return f"1 {singular}"
    return f"{count} {plural or (singular + 's')}"
