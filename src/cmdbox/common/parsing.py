_UNITS = {
    "b": 1,
    "kb": 1024,
    "mb": 1024**2,
    "gb": 1024**3,
}


def parse_byte_size(value: str | int) -> int:
    if isinstance(value, int):
        result = value  # Already bytes, pass through
    else:
        cleaned = value.strip().lower().replace(" ", "")
        for suffix, multiplier in sorted(_UNITS.items(), key=lambda x: -len(x[0])):
            if cleaned.endswith(suffix):
                number = cleaned[: -len(suffix)]
                try:
                    result = int(float(number) * multiplier)
                    break
                except ValueError:
                    raise ValueError(f"Invalid byte size: '{value}'")
        else:
            # If no unit, treat as raw bytes
            try:
                result = int(cleaned)
            except ValueError:
                raise ValueError(f"Invalid byte size: '{value}'")
    if result < 0:
        # A negative size is never valid (e.g. it would silently disable
        # log-file rotation downstream).
        raise ValueError(f"Invalid byte size: '{value}'")
    return result
