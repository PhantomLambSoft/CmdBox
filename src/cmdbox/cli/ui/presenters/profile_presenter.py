from typing import Sequence, Callable

from cmdbox.cli.ui.primitives import col, pluralize, table_panel, kv_table, section
from cmdbox.models import Profile
from cmdbox.services.profile_services import ProfileStatus

PROFILE_COLUMNS: dict[str, tuple[str, dict, Callable[[Profile], object]]] = {
    "name": (
        "Name",
        {"style": "entity.name", "no_wrap": True},
        lambda p: p.name,
    ),
    "description": (
        "Description",
        {"overflow": "fold"},
        lambda p: p.description,
    ),
    "date_created": (
        "Created",
        {"style": "entity.time", "no_wrap": True},
        lambda p: p.date_created,
    ),
    "last_used": (
        "Last Used",
        {"style": "entity.time", "no_wrap": True},
        lambda p: p.last_used,
    ),
}


DEFAULT_FIELDS = ["name", "description", "last_used"]


def render_profile_created(profile: Profile):
    rendered_profile = render_profile(profile)
    return section(
        title=f"Profile '{profile.name}' created",
        body=rendered_profile,
        border_style="status.success",
    )


def render_profile(profile: Profile):
    rows = []
    for value in PROFILE_COLUMNS.values():
        header, _, extractor = value
        rows.append((header, extractor(profile)))
    return kv_table(rows)


def render_profile_list(
    profiles: Sequence[Profile], *, title: str = None, fields: list[str] | None = None
):
    active_fields = fields or DEFAULT_FIELDS
    active_fields = [f for f in active_fields if f in PROFILE_COLUMNS]

    columns = []
    extractors = []

    for field in active_fields:
        header, col_args, extractor = PROFILE_COLUMNS[field]
        columns.append(col(header, **col_args))
        extractors.append(extractor)

    rows = [tuple(extractor(p) for extractor in extractors) for p in profiles]

    caption = f"{pluralize(len(profiles), 'profile')} found"

    return table_panel(
        title=title or "Profiles",
        columns=columns,
        rows=rows,
        caption=caption,
    )


def render_profile_updated(profile: Profile):
    rendered_profile = render_profile(profile)
    return section(
        title=f"Profile '{profile.name}' updated",
        body=rendered_profile,
        border_style="status.success",
    )


def render_profile_deleted(profile: Profile):
    rendered_profile = render_profile(profile)
    return section(
        title=f"Profile '{profile.name}' deleted",
        body=rendered_profile,
        border_style="status.success",
    )


def render_profile_switched(profile: Profile, *, scope: str):
    return section(
        title=f"Switched {scope} profile to '{profile.name}'",
        body=render_profile(profile),
        border_style="status.success",
    )


def render_profile_status(status: ProfileStatus):
    rows = [
        ("Command profile", status.command_profile),
        ("Variable profile", status.variable_profile),
        ("Settings profile", status.settings_profile),
        ("Linked", "Yes" if status.linked else "No"),
    ]
    return kv_table(rows)
