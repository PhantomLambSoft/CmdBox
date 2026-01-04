from cmdbox.cli.prompts.completers import TagCompleter
from cmdbox.cli.prompts.prompts import prompt_for_tags
from cmdbox.cli.prompts.validators import TagNameValidator
from cmdbox.services.tag_services import TagServices


def get_tags_interactive(tag_services: TagServices) -> list[str] | None:
    def get_tags(query: str) -> list[str]:
        found_tags = tag_services.search(query, fields="name")
        return [tag.name for tag in found_tags]

    tag_completer = TagCompleter(get_tags)
    validator = TagNameValidator()
    tags = prompt_for_tags(tag_completer, validator)
    return tags
