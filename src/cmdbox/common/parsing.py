_UNITS = {
    "b": 1,
    "kb": 1024,
    "mb": 1024**2,
    "gb": 1024**3,
}


def parse_byte_size(value: str | int) -> int:
    if isinstance(value, int):
        return value  # Already bytes, pass through
    cleaned = value.strip().lower().replace(" ", "")
    for suffix, multiplier in sorted(_UNITS.items(), key=lambda x: -len(x[0])):
        if cleaned.endswith(suffix):
            number = cleaned[: -len(suffix)]
            try:
                return int(float(number) * multiplier)
            except ValueError:
                raise ValueError(f"Invalid byte size: '{value}'")
    # If no unit, treat as raw bytes
    try:
        return int(cleaned)
    except ValueError:
        raise ValueError(f"Invalid byte size: '{value}'")
