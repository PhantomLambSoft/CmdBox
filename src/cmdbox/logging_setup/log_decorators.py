import logging
import time
from functools import wraps
from typing import Callable, Any


def log_action(
    module_name: str,
    action_name: str,
):
    """
    Decorator to log the execution of a function, including the start, end, and elapsed time in milliseconds.
    Logs will be captured using the specified module name and include details about the specified action name.

    The decorator also logs exceptions if the wrapped function raises one.

    Args:
        module_name (str): The name of the module to associate with the logger.
        action_name (str): The name of the action to log.
    """

    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            log = logging.getLogger(module_name)
            start_time = time.perf_counter()
            log.info("%s start", action_name)

            try:
                result = func(*args, **kwargs)
                return result
            except Exception:
                log.exception("%s failed", action_name)
                raise
            finally:
                elapsed_ms = (time.perf_counter() - start_time) * 1000
                log.info("%s finished in %.2f ms", action_name, elapsed_ms)

        return wrapper

    return decorator
