import os
from pathlib import Path
from peewee import SqliteDatabase

DB_PATH = Path.home() / ".cmdbox" / "cmdbox.db"
DB_PATH.parent.mkdir(parents=True, exist_ok=True)


db = SqliteDatabase(None)


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
    if testing is None:
        env = os.environ.get("CMDBOX_ENV", "production")
        testing = env == "testing"

    if testing:
        db_path = ":memory:"
    else:
        db_path = str(DB_PATH)
    db.init(db_path)


def init_production_db(models: list[type]) -> None:
    init_database(False)
    db.connect(reuse_if_open=True)
    db.create_tables(models)
