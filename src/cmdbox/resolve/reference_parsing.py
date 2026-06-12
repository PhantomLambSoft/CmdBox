from typing import Optional

from .type_defs import RefKind


def read_angle_token(s: str, start_i: int) -> tuple[Optional[str], str, int]:
    """
    Reads a token beginning with "<" until an unescaped ">" is encountered.

    Args:
        s (str): The string containing the token.
        start_i (int): The index of the first character of the token.

    Returns:
        str: token_inner (with escapes already interpreted), or None if no closing bracket.
        str: raw_token - the original token string, including the angle brackets.
        int: next_i - the index of the first character after the closing bracket.

    """
    assert s[start_i] == "<"
    i = start_i + 1
    n = len(s)

    inner_chars: list[str] = []

    while i < n:
        ch = s[i]
        if ch == "\\":
            if i + 1 < n and s[i + 1] in ("\\", "<", ">"):
                inner_chars.append(s[i + 1])
                i += 2
                continue
            inner_chars.append("\\")
            i += 1
            continue

        if ch == ">":
            raw_token = s[start_i + 1 : i]
            token_inner = "".join(inner_chars).strip()
            return token_inner, raw_token, i + 1

        inner_chars.append(ch)
        i += 1

    return None, s[start_i:], n


def parse_kind_and_key(token_inner: str) -> tuple[RefKind, str]:
    """
    Parses the kind and key from a given token string.

    This method evaluates the provided token and determines its type, which can be a
    command or variable, based on the prefix present in the token. If no prefix is
    found, it defaults to `RefKind.VARIABLE`. The method ensures that parts of the
    token are appropriately stripped of whitespace during processing.

    Args:
        token_inner (str): The input token string that may specify a prefix and a key,
            separated by a colon.

    Returns:
        tuple[RefKind, str]: A tuple containing the kind of reference (`RefKind`) and
        its corresponding key as a string.
    """
    if ":" not in token_inner:
        return RefKind.VARIABLE, token_inner
    prefix, key = (part.strip() for part in token_inner.split(":", 1))

    if prefix == "cmd":
        return RefKind.COMMAND, key
    if prefix == "var":
        return RefKind.VARIABLE, key
    return RefKind.VARIABLE, token_inner


def extract_references(template: str) -> list[tuple[RefKind, str]]:
    """
    Extracts references from a given template string.

    Parses through the input template string to identify specific patterns enclosed
    in angle brackets (`<` and `>`). Each such identified token is processed to
    extract its kind and key, which are then appended to a list of references
    represented as tuples.

    Args:
        template (str): The input string containing the template to be parsed.

    Returns:
        list[tuple[RefKind, str]]: A list of tuples where each tuple contains the
        kind and key extracted from the references found within the template.
    """
    refs: list[tuple[RefKind, str]] = []
    i = 0
    n = len(template)

    while i < n:
        ch = template[i]
        if ch == "\\":
            i += 2
            continue
        if ch == "<":
            token_inner, _, next_i = read_angle_token(template, i)
            if token_inner:
                refs.append(parse_kind_and_key(token_inner))
            i = next_i
            continue
        i += 1

    return refs
