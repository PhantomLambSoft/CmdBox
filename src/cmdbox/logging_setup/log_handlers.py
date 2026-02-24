import sys
import re
import logging
from logging.handlers import RotatingFileHandler

from cmdbox.logging_setup.log_config import LogConfig, get_logger


class SecretRedactionFilter(logging.Filter):
    """
    A logging filter to redact sensitive information from log messages.

    This filter is designed to sanitize log messages by redacting sensitive
    data such as tokens and passwords. The redaction process is applied
    whenever a log record passes through the filter, ensuring that such
    sensitive information is not exposed in logs.
    """

    def filter(self, record: logging.LogRecord) -> bool:
        if isinstance(record.msg, str):
            record.msg = self._redact(record.msg)
        return True

    def _redact(self, s: str) -> str:
        s = re.sub(r"(?i)\b(token\s*=\s*)(\S+)", r"\1[REDACTED]", s)
        s = re.sub(r"(?i)\b(password\s*=\s*)(\S+)", r"\1[REDACTED]", s)
        return s


def configure_logging(config: LogConfig) -> None:
    """
    Configures logging for the application, setting up a logger with both console and file
    handlers as specified by the provided configuration. Filters and formatters are applied
    to ensure appropriate logging output.

    Args:
        config (LogConfig): Configuration object specifying the logging settings, including
            log levels, file path, maximum file size, and number of backups.
    """
    logger = get_logger()
    logger.setLevel(logging.DEBUG)  # capture everything, then filter
    logger.propagate = False

    for h in list(logger.handlers):
        logger.removeHandler(h)

    formatter = logging.Formatter(
        fmt="%(asctime)s | %(levelname)s | %(name)s: %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    # Console handler
    ch = logging.StreamHandler(sys.stderr)
    ch.setLevel(config.console_level)
    ch.setFormatter(formatter)
    ch.addFilter(SecretRedactionFilter())
    logger.addHandler(ch)

    # File handler
    if config.file_enabled:
        fh = RotatingFileHandler(
            filename=str(config.file_path),
            maxBytes=int(config.max_bytes),
            backupCount=int(config.backups),
            encoding="utf-8",
        )
        fh.setLevel(config.file_level)
        fh.setFormatter(formatter)
        fh.addFilter(SecretRedactionFilter())
        logger.addHandler(fh)
