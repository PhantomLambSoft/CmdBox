from dataclasses import dataclass
from typing import Callable, Sequence, Iterable, Tuple

from prompt_toolkit.completion import Completer, Completion, CompleteEvent
from prompt_toolkit.document import Document


def _normalize_tag(s: str) -> str:
    """
    Normalizes a tag by stripping leading and trailing whitespace and converting the string to lowercase.

    Args:
        s (str): The string to be normalized.

    Returns:
        str: The normalized string.
    """
    return s.strip().lower()


def _split_csv_like(text: str) -> Tuple[str, str, int]:
    """
    Split input into:
        - prefix: Everything up to and including the last comma (and any following spaces)
        - active: The chunk the user is currently typing (no comma)
        - start_pos: The start index of active within the full text
    Args:
        text:

    Returns:

    """
    last_comma = text.rfind(",")
    if last_comma == -1:
        prefix = ""
        active = text
        start_pos = 0
        return prefix, active, start_pos

    prefix_end = last_comma + 1
    prefix = text[:prefix_end]
    i = prefix_end
    while i < len(text) and text[i].isspace():
        i += 1
    prefix = text[:i]
    active = text[i:]
    start_pos = i
    return prefix, active, start_pos


def _simple_fuzzy_score(needle: str, haystack: str) -> int:
    """
    Lightweight fuzzy score with no external dependencies.
    Higher is better.  Range is 0...100.
    Args:
        needle: The search term to match within the haystack
        haystack: The text to search for needle within

    Returns:
        An integer score indicating how well needle matches haystack
    """
    if not needle:
        return 1
    if haystack.startswith(needle):
        return 100
    if needle in haystack:
        return 75
    it = iter(haystack)
    if all(ch in it for ch in needle):
        return 55
    return 0


@dataclass
class TagCompleterConfig:
    min_score: int = 1
    max_results: int = 12
    case_insensitive: bool = True


class TagCompleter(Completer):
    """
    Implements a tag completion mechanism for input fields.

    This class provides dynamic suggestions for completing partial user inputs based on a
    predefined pool of tags. The suggestions are context-aware, configurable, and sorted
    by relevance. It is specifically designed for scenarios where inputs consist of a
    comma-separated list of tags.

    Attributes:
        _get_tags (Callable[[], Sequence[str]]): A function that retrieves the list of
            available tags.
        _config (TagCompleterConfig): Configuration object that determines completion
            behavior, including case sensitivity, scoring thresholds, and maximum results.
    """

    def __init__(
        self,
        get_tags: Callable[[], Sequence[str]],
        config: TagCompleterConfig | None = None,
    ):
        self._get_tags = get_tags
        self._config = config if config is not None else TagCompleterConfig()

    def get_completions(
        self, document: Document, complete_event: CompleteEvent
    ) -> Iterable[Completion]:
        """
        Generate completions for a partial text input based on a list of pre-defined tags.

        This method processes the input text before the cursor, normalizes it based on
        configuration settings, and evaluates potential completions from a pool of tags.
        Completions are filtered and sorted by relevance, and only unique, unused tags
        meeting a minimum score threshold are suggested.

        Args:
            document (Document): The input context containing the text typed by the user.
            complete_event (CompleteEvent): Event instance providing additional
                information about the completion trigger.

        Yields:
            Completion: A generator of Completion objects, each containing a potential
                completion string, its display text, and the position within the text
                where it should be applied.
        """
        text = document.text_before_cursor
        prefix, active, active_start = _split_csv_like(text)

        active_norm = (
            _normalize_tag(active) if self._config.case_insensitive else active.strip()
        )
        if self._config.case_insensitive:
            tag_pool = [(t, _normalize_tag(t)) for t in self._get_tags()]
        else:
            tag_pool = [(t, t.strip()) for t in self._get_tags()]

        # Do not suggest tags already used in the list
        already_raw = prefix.split(",")
        already = {_normalize_tag(t) for t in already_raw if t.strip()}
        scored: list[Tuple[int, str]] = []

        for original, norm in tag_pool:
            if not norm:
                continue
            if norm in already:
                continue
            score = _simple_fuzzy_score(active_norm, norm)
            if score >= self._config.min_score:
                scored.append((score, original))

        scored.sort(key=lambda x: (-x[0], x[1].lower()))
        # Replace only the active chunk, not the entire line
        # Negative chars to delete before inserting completion
        start_pos = -len(active)

        for score, tag in scored[: self._config.max_results]:
            yield Completion(
                text=tag,
                start_position=start_pos,
                display=tag,
            )
