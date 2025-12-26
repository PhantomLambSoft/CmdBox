from typing import Sequence

from peewee import IntegrityError

from .base_repository import BaseRepository
from .errors import NameConflictError, UnknownNameError, ValidationError
from .validators import TagValidator
from cmdbox.models import Tag


class TagRepository(BaseRepository[Tag]):
    model = Tag

    def __init__(self, validator: TagValidator | None = None):
        self.validator = validator or TagValidator()

    def create(self, name: str, description: str | None = None) -> Tag:
        name = name.strip() if name else None
        self.validator.validate_create(name=name, description=description)
        try:
            return Tag.create(name=name, description=description)
        except IntegrityError as exc:
            if name is not None and self._is_unique_name_violation(exc):
                raise NameConflictError(name=name) from exc
            raise

    def get_by_name(self, name: str) -> Tag | None:
        name = name.lower()
        tag = Tag.get_or_none(Tag.name == name)
        if tag is None:
            raise UnknownNameError(name=name)
        return tag

    def update(self, tag_name: str, **fields) -> Tag | None:
        tag = self.get_by_name(tag_name)
        if not fields:
            raise ValueError("No fields provided for update.")

        if "name" in fields and fields.get("name") is not None:
            fields["name"] = fields.get("name").strip()

        self.validator.validate_update(
            name=fields.get("name", tag.name),
            description=fields.get("description", tag.description),
        )
        try:
            for key, value in fields.items():
                if not hasattr(tag, key):
                    raise ValidationError(f"Invalid field: {key}")
                if value is not None:
                    setattr(tag, key, value)
            tag.save()
            return tag
        except IntegrityError as exc:
            name = fields.get("name", "")
            if name is not None and self._is_unique_name_violation(exc):
                raise NameConflictError(name=name) from exc
            raise

    def list_all(
        self, order_by: str | Sequence[str] = "name", limit: int = 25
    ) -> list[Tag]:
        ordering = self._resolve_ordering(order_by)
        return list(Tag.select().order_by(*ordering).limit(limit))

    def search(
        self, query: str, fields: str | Sequence[str] | None = ("name", "description")
    ) -> list[Tag]:
        return self._search(query, "name", fields)

    def delete(self, name: str) -> bool:
        tag = self.get_by_name(name)
        if not tag:
            return False
        tag.delete_instance()
        return True
