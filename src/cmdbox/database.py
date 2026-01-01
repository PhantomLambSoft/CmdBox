import os
from peewee import SqliteDatabase

from cmdbox.core.paths import get_app_data_dir

DB_PATH = get_app_data_dir() / "cmdbox.db"
DB_PATH.parent.mkdir(parents=True, exist_ok=True)


db = SqliteDatabase(None)
_db_initialized = False


def init_database(testing: bool = False) -> None:
    """
    Initializes the database connection based on the environment or testing flag.

    This function determines whether to use an in-memory SQLite database for
    testing purposes or the standard SQLite database file path for production
    or other environments. The behavior is controlled by the `testing` argument
    and the `CMDBOX_ENV` environment variable.

    Args:
        testing (bool | None): A flag to indicate whether to use the testing
            (in-memory) database. If None, the value is derived from the
            `CMDBOX_ENV` environment variable. Defaults to None.
    """
    global _db_initialized
    if _db_initialized:
        return
    if testing is None:
        env = os.environ.get("CMDBOX_ENV", "production")
        testing = env == "testing"

    if testing:
        db_path = ":memory:"
    else:
        db_path = str(DB_PATH)
    db.init(db_path)
    _db_initialized = True


def get_db(testing: bool = False) -> SqliteDatabase:
    init_database(testing)
    if db.is_closed():
        db.connect(reuse_if_open=True)
    return db


def ensure_schema() -> None:
    from cmdbox.models import ALL_MODELS

    _db = get_db()
    _db.create_tables(ALL_MODELS, safe=True)
