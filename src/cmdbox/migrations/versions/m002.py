"""
v2: Adds named profile support. Adds Profile and ProfileState tables
and adds a 'profile' foreign key to Command and Variable tables. Existing
commands and variables are backfilled onto a new "default" profile.
"""

from peewee import SqliteDatabase
from datetime import datetime

VERSION = 2


def migrate(old_db_path: str, new_db_path: str) -> None:
    old_db = SqliteDatabase(old_db_path)
    new_db = SqliteDatabase(new_db_path)

    old_db.connect()
    new_db.connect()

    try:
        from cmdbox.models import (
            Profile,
            ProfileState,
            Command,
            Variable,
            Tag,
            CommandTag,
            VariableTag,
            CommandHistory,
        )

        new_models = [
            Profile,
            ProfileState,
            Command,
            Variable,
            Tag,
            CommandTag,
            VariableTag,
            CommandHistory,
        ]

        with new_db.bind_ctx(new_models):
            new_db.create_tables(new_models)

            default_profile = Profile.create(
                name="default",
                description="Automatically created default profile.",
                date_created=datetime.now(),
                last_used=None,
            )

            new_db.execute_sql("ATTACH DATABASE ? AS old", (old_db_path,))
            try:
                # Unchanged tables copy directly as is
                new_db.execute_sql(
                    "INSERT INTO tag (id, name, description, date_created, last_updated) "
                    "SELECT id, name, description, date_created, last_updated FROM old.tag"
                )
                new_db.execute_sql(
                    "INSERT INTO commandtag (id, command_id, tag_id, date_created) "
                    "SELECT id, command_id, tag_id, date_created FROM old.commandtag"
                )
                new_db.execute_sql(
                    "INSERT INTO variabletag (id, variable_id, tag_id, date_created) "
                    "SELECT id, variable_id, tag_id, date_created FROM old.variabletag"
                )
                new_db.execute_sql(
                    "INSERT INTO command_history "
                    "(id, alias, template, resolved, variables_used, exit_code, ran_at) "
                    "SELECT id, alias, template, resolved, variables_used, exit_code, ran_at "
                    "FROM old.command_history"
                )

                # Command and variable pick up the new profile_id column, set
                # to the default profile created above.
                new_db.execute_sql(
                    "INSERT INTO command "
                    "(id, alias, template, description, cwd, shell, env, timeout, "
                    "date_created, last_updated, used, last_used, profile_id) "
                    "SELECT id, alias, template, description, cwd, shell, env, timeout, "
                    "date_created, last_updated, used, last_used, ? FROM old.command",
                    (default_profile.id,),
                )
                new_db.execute_sql(
                    "INSERT INTO variable "
                    "(id, name, value, date_created, last_updated, profile_id) "
                    "SELECT id, name, value, date_created, last_updated, ? FROM old.variable",
                    (default_profile.id,),
                )
                new_db.execute_sql(
                    "INSERT INTO command_history "
                    "(id, alias, template, resolved, variables_used, exit_code, ran_at, profile_id) "
                    "SELECT id, alias, template, resolved, variables_used, exit_code, ran_at, ? "
                    "FROM old.command_history",
                    (default_profile.id,),
                )
            finally:
                new_db.execute_sql("DETACH DATABASE old")

            ProfileState.create(
                active_command_profile=default_profile,
                active_variable_profile=default_profile,
                active_settings_profile=default_profile,
            )

    finally:
        old_db.close()
        new_db.close()
