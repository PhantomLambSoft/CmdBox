from prompt_toolkit.document import Document
from prompt_toolkit.validation import Validator, ValidationError


class AliasValidator(Validator):

    def validate(self, document: Document) -> None:
        text = document.text.strip()
        if text == "":
            raise ValidationError(message="Alias cannot be empty")
        if " " in text:
            raise ValidationError(message="Alias cannot contain spaces")
        if "<" in text or ">" in text:
            raise ValidationError(message="Alias cannot contain '<' or '>' characters")


class TemplateValidator(Validator):

    def validate(self, document: Document) -> None:
        text = document.text.strip()
        if text == "":
            raise ValidationError(message="Template cannot be empty")


class NameValidator(Validator):

    def validate(self, document: Document) -> None:
        text = document.text.strip()
        if text == "":
            raise ValidationError(message="Name cannot be empty")
        if " " in text:
            raise ValidationError(message="Name cannot contain spaces")
        if "<" in text or ">" in text:
            raise ValidationError(message="Name cannot contain '<' or '>' characters")
