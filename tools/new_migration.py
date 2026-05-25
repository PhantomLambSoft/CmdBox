"""
Developer tool: scaffolds the next migration file in src/cmdbox/migrations/.

Usage:
    python tools/new_migration.py
"""

import re
import sys
from pathlib import Path

MIGRATIONS_DIR = (
    Path(__file__).parent.parent / "src" / "cmdbox" / "migrations" / "versions"
)

TEMPLATE = '''
"""
v{version}: TODO - describe what this migration changes in the database.
"""
from peewee import SqliteDatabase

VERSION = {version}

def migrate(old_db_path: str, new_db_path: str) -> None:
    old_db = SqliteDatabase(old_db_path)
    new_db = SqliteDatabase(new_db_path)
    
    old_db.connect()
    new_db.connect()
    
    try:
        # TODO: create table in new_db with updated schema
        # TODO: copy and transform data from old_db into new_db
        pass
    finally:
        old_db.close()
        new_db.close()
    '''


def get_next_version() -> int:
    existing = sorted(MIGRATIONS_DIR.glob("m[0-9]*.py"))
    if not existing:
        return 2  # v1 is the baseline. The first real migration will be v2
    last = existing[-1]
    match = re.search(r"m(\d+)\.py$", last.name)
    if not match:
        print(f"Could not parse version number from: {last.name}", file=sys.stderr)
        sys.exit(1)
    return int(match.group(1)) + 1


def main() -> None:
    next_version = get_next_version()
    filename = f"m{next_version:03d}.py"
    path = MIGRATIONS_DIR / filename

    if path.exists():
        print(f'File already exists: "{path}"', file=sys.stderr)
        sys.exit(1)

    path.write_text(TEMPLATE.format(version=next_version), encoding="utf-8")

    print(f"Created: {path}")
    print(f"  --> Remember to update CURRENT_VERSION in runner.py to {next_version}")


if __name__ == "__main__":
    main()
