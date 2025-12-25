from dataclasses import dataclass
from typing import List


@dataclass
class TagAttachResult:
    """
    Represents the result of attaching tags.

    This class holds information about successfully added tags and tags
    that were already present. It helps in distinguishing between newly
    added tags and pre-existing tags when managing tag assignments.

    Attributes:
        added (List[str]): List of tags that were newly added.
        existing (List[str]): List of tags that were already present.
    """

    added: List[str]
    existing: List[str]


@dataclass
class TagDetachResult:
    """
    Represents the result of detaching tags.

    This class holds information about tags that were successfully removed
    and tags that were not attached in the first place to be removed.

    Attributes:
        removed (List[str]): List of tags that were successfully removed.
        not_attached (List[str]): List of tags that were not attached in the first place.
    """

    removed: List[str]
    not_attached: List[str]
