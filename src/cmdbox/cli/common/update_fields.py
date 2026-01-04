from typing import Any, Optional, Iterable, Dict
import typer


def parse_set_pairs(pairs: Optional[list[str]]) -> Dict[str, str]:
    """
    Parses a list of key-value string pairs and converts them into a dictionary.

    This function takes a list of strings formatted as "key=value", validates their structure,
    and converts them into a dictionary where the keys and values correspond to the
    parsed components of the strings. Invalid pairs result in an error being raised.

    Args:
        pairs (Optional[list[str]]): A list of strings where each string is formatted
            as "key=value". If the list is empty or None, an empty dictionary is returned.

    Returns:
        Dict[str, str]: A dictionary where the keys and values correspond to the parsed
            key-value pairs from the input list.

    Raises:
        typer.BadParameter: If any string in the list does not contain an "=" character,
            or if the key part is empty after splitting.
    """
    out: Dict[str, str] = {}
    if not pairs:
        return out
    for item in pairs:
        if "=" not in item:
            raise typer.BadParameter(f"Invalid --set value '{item}'. Use key=value.")
        k, v = item.split("=", 1)
        k = k.strip()
        if not k:
            raise typer.BadParameter(
                f"Invalid --set value '{item}'. Key cannot be empty."
            )
        out[k] = v
    return out


def merge_fields(base: Dict[str, Any], extra: Dict[str, Any]) -> Dict[str, Any]:
    """
    Merges two dictionaries and ensures there are no overlapping keys. If overlapping keys
    are detected, an error is raised with a detailed message indicating the conflicts.

    Args:
        base (Dict[str, Any]): The base dictionary to merge into.
        extra (Dict[str, Any]): The additional dictionary containing keys and values to
            merge into the base dictionary.

    Returns:
        Dict[str, Any]: A new dictionary with the combined keys and values from the
        `base` and `extra` dictionaries.

    Raises:
        typer.BadParameter: If the `base` and `extra` dictionaries have overlapping keys.
    """
    conflicts = set(base).intersection(extra)
    if conflicts:
        keys = ", ".join(sorted(conflicts))
        raise typer.BadParameter(f"Field(s) specified multiple ways: {keys}")
    merged = dict(base)
    merged.update(extra)
    return merged


def filter_allowed(fields: Dict[str, Any], allowed: Iterable[str]) -> Dict[str, Any]:
    """
    Filters a dictionary to retain only allowed keys and raises an error if unknown keys are
    present in the dictionary.

    Args:
        fields (Dict[str, Any]): A dictionary containing the fields to filter.
        allowed (Iterable[str]): An iterable of allowed keys.

    Returns:
        Dict[str, Any]: A dictionary containing only the fields with keys found in the allowed
        set.

    Raises:
        typer.BadParameter: If the dictionary contains keys not found in the allowed set.
    """
    allowed_set = set(allowed)
    unknown = [k for k in fields.keys() if k not in allowed_set]
    if unknown:
        keys = ", ".join(sorted(unknown))
        raise typer.BadParameter(f"Unknown field(s) for --set: {keys}")
    return fields
