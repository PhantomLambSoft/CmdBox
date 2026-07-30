from cmdbox.models import Profile, Command, Variable, Tag


def profile_field_options(incomplete: str) -> list[str]:
    fields = Profile._meta.sorted_field_names
    return [field for field in fields if field.startswith(incomplete)]


def command_field_options(incomplete: str) -> list[str]:
    fields = Command._meta.sorted_field_names
    return [field for field in fields if field.startswith(incomplete)]


def command_editable_field_options(incomplete: str) -> list[str]:
    fields = ["alias", "template", "description"]
    return [field for field in fields if field.startswith(incomplete)]


def variable_field_options(incomplete: str) -> list[str]:
    fields = Variable._meta.sorted_field_names
    return [field for field in fields if field.startswith(incomplete)]


def variable_editable_field_options(incomplete: str) -> list[str]:
    fields = ["name", "value"]
    return [field for field in fields if field.startswith(incomplete)]


def tag_field_options(incomplete: str) -> list[str]:
    fields = Tag._meta.sorted_field_names
    return [field for field in fields if field.startswith(incomplete)]


def tag_editable_field_options(incomplete: str) -> list[str]:
    fields = ["name", "description"]
    return [field for field in fields if field.startswith(incomplete)]
