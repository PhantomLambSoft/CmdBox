# CmdBox Repository Layout

This document describes the intended layout of the CmdBox repository.

- Product: CmdBox
- Python package: `cmdbox`
- CLI binary: `cb`
- Persistence: SQLite using Peewee models

---

## Top level layout

```text
CmdBox/
├─ pyproject.toml
├─ README.md
├─ LICENSE
├─ .gitignore
├─ docs/
│  ├─ cmdbox_design.md          # Core behavior and resolution rules
│  ├─ cb_cli_contract.md        # CLI contract and public surface
│  ├─ architecture_overview.md  # High level architecture, data flow
│  ├─ repository_layout.md      # This document
│  └─ roadmap.md                # Phased development plan
├─ src/
│  └─ cmdbox/
│     ├─ __init__.py
│     ├─ __main__.py            # Allows `python -m cmdbox` to run the CLI
│     ├─ config.py              # Config loading and defaults (db path, etc.)
│     ├─ db.py                  # SQLite database object and init helpers
│     ├─ models.py              # Peewee Model subclasses: Command, Variable
│     ├─ repositories.py        # Optional repository layer over Peewee models
│     ├─ errors.py              # Error types (UnknownAlias, UnknownVariable, etc.)
│     ├─ resolver.py            # Core engine for interpolation and resolution
│     ├─ execution/
│     │  ├─ __init__.py
│     │  └─ executor.py         # Executor abstraction for running resolved commands
│     ├─ cli/
│     │  ├─ __init__.py
│     │  ├─ main.py             # Top level CLI app, entry point for `cb`
│     │  ├─ commands.py         # Command management subcommands (list, add, run, etc.)
│     │  └─ vars.py             # Variable management subcommands
│     └─ logging_setup.py       # Optional simple logging setup for debug flags
└─ tests/
   ├─ __init__.py
   ├─ test_db_models.py           # SQLite + Peewee model tests
   ├─ test_repositories.py        # Repository behavior tests (if using repositories)
   ├─ test_resolver_core.py       # Core resolver tests
   ├─ test_execution_executor.py  # Execution abstraction tests
   └─ test_cli/
      ├─ __init__.py
      ├─ test_cli_commands.py     # Tests for `cb add`, `cb list`, `cb show`, `cb run`
      └─ test_cli_vars.py         # Tests for `cb vars list`, `cb vars add`, etc.
````

---

## 1. Top level files

* `pyproject.toml`

    * Project metadata and dependency definitions.
    * Defines a console script for the CLI, for example:

      ```toml
      [project.scripts]
      cb = "cmdbox.cli.main:app"
      ```

      or whatever entry point your chosen CLI framework requires.

* `README.md`

    * Project overview.
    * Quick start with `cb vars add`, `cb add`, `cb run`.

* `LICENSE`

    * License for CmdBox.

* `.gitignore`

    * Git ignore rules for Python, SQLite files, editor junk, and local artifacts.

* `docs/`

    * Documentation and design files:

        * Behavior spec (`cmdbox_design.md`)
        * CLI contract (`cb_cli_contract.md`)
        * Architecture overview
        * Repository layout
        * Roadmap

---

## 2. `src/cmdbox` package

This is the main CmdBox library. The goal is to keep responsibilities separated but still easy to navigate.

### 2.1 `config.py`

* Handles configuration related to:

    * Database file path.
    * Optional config file in the user home directory.
    * Environment variable overrides.
* Provides simple functions like:

    * `get_database_path()`
    * `load_config()`

### 2.2 `db.py`

* Owns the Peewee `SqliteDatabase` instance and database initialization logic.

Typical contents:

* A global `db = SqliteDatabase(path)` that is configured at runtime.
* An `init_db()` function that:

    * Resolves the database path via `config.py`.
    * Binds models to the database.
    * Creates tables if they do not exist.

Example responsibilities:

* Opening and closing connections.
* Ensuring `Command` and `Variable` tables are created.

### 2.3 `models.py`

* Contains Peewee `Model` subclasses representing persistent entities.

Core models:

* `Command`

    * Fields like:

        * `alias` (unique)
        * `template`
        * `description` (nullable)
        * `tags` (for MVP can be a simple text field storing comma separated tags)

* `Variable`

    * Fields like:

        * `name` (unique)
        * `value`

These models are used both by the resolver (as data carriers) and by the CLI and repositories for persistence.

### 2.4 `repositories.py`

* Optional but useful thin layer encapsulating common database operations.

For example:

* `CommandRepository`

    * `list()`
    * `get_by_alias(alias)`
    * `create_or_update(...)`
    * `delete(alias)`

* `VariableRepository`

    * `list()`
    * `get_by_name(name)`
    * `set_value(name, value)`
    * `delete(name)`

This provides a focused interface that the resolver and CLI can rely on, while the implementation uses Peewee behind the
scenes.

### 2.5 `errors.py`

* All custom exception types live here to keep them consistent.

Examples:

* `CbError` (base class)
* `UnknownAlias`
* `UnknownVariable`
* `UnknownCommandReference`
* `ResolutionError`
* `DatabaseError` (if you decide to wrap Peewee exceptions)

### 2.6 `resolver.py`

* Contains the pure resolution logic:

Responsibilities:

* Parse and expand `<name>` variable tokens.
* Parse and expand `<cmd:alias>` command reference tokens.
* Detect and handle cycles.
* Raise well defined errors for missing references.

Key idea:

* Even though it uses Peewee model instances, treat them as plain objects with attributes.
* It should not import or configure the database directly.
* It should not run queries itself when possible. Queries should be handled in repositories or passed data.

You can structure it as:

* A low level function that takes plain mappings or lists of commands and variables.
* A higher level helper that uses repositories to load what is needed and calls the low level function.

### 2.7 `execution/`

* Execution logic is isolated here.

`execution/executor.py`:

* Defines an `Executor` class or simple functions for:

    * `run(command: str) -> int`

        * MVP: uses `subprocess.run(command, shell=True)`.

* Optionally extra helpers later:

    * Dry run support (return without executing).
    * Capturing stdout or stderr.

The resolver never executes commands. The CLI decides if and when to call the executor.

### 2.8 `cli/`

* Everything related to the `cb` command line interface.

`cli/main.py`:

* The entry point for the CLI application.
* Sets up the CLI framework (Click, Typer, argparse, etc.).
* Registers subcommands:

    * `list`, `show`, `add`, `edit`, `remove`, `run`
    * `vars list`, `vars show`, `vars add`, `vars edit`, `vars remove`
* Handles global options like `--version`, `--config`, `--store`, `--json`, `--debug`.

`cli/commands.py`:

* Implementation of command management subcommands, for example:

    * `cb list`
    * `cb show <alias>`
    * `cb add <alias> --template ...`
    * `cb edit <alias> ...`
    * `cb remove <alias>`

* These functions:

    * Initialize the database.
    * Use repositories or Peewee models for data access.
    * Invoke the resolver when needed.
    * Format and print output.

`cli/vars.py`:

* Implementation of `cb vars` subcommands, for example:

    * `cb vars list`
    * `cb vars show <name>`
    * `cb vars add <name> --value ...`
    * `cb vars edit <name> --value ...`
    * `cb vars remove <name>`

### 2.9 `logging_setup.py`

* Optional module for configuring logging:

    * Default log level.
    * Behavior when `--debug` is passed to the CLI.
    * Where logs are written (stdout only for MVP, or optional file).

### 2.10 `__main__.py`

* Allows `python -m cmdbox` to behave like calling `cb`.

Example:

```python
from cmdbox.cli.main import app

if __name__ == "__main__":
    app()
```

The exact details depend on your CLI framework, but the pattern is the same.

---

## 3. `tests` package

Tests mirror the structure of `src/cmdbox` so it is easy to find where to add new tests.

* `test_db_models.py`

    * Tests that Peewee models:

        * Create tables.
        * Enforce uniqueness.
        * Store and retrieve values correctly.
    * Uses a temporary SQLite file or in memory database.

* `test_repositories.py`

    * Tests each method of `CommandRepository` and `VariableRepository`.
    * Ensures consistent behavior in edge cases.

* `test_resolver_core.py`

    * Tests the resolver against the examples and rules in `cmdbox_design.md`.
    * Covers:

        * Variable interpolation.
        * Nested variables.
        * Command references.
        * Cycle detection.
        * Missing references.

* `test_execution_executor.py`

    * Tests that `Executor.run`:

        * Returns expected exit codes.
        * Handles simple commands and error cases.

* `test_cli/`

    * `test_cli_commands.py`:

        * Simulates CLI usage for:

            * `cb add`
            * `cb list`
            * `cb show`
            * `cb remove`
            * `cb run` (dry run and real run, as much as practical in tests)

    * `test_cli_vars.py`:

        * Simulates:

            * `cb vars add`
            * `cb vars list`
            * `cb vars show`
            * `cb vars edit`
            * `cb vars remove`

Tests should keep the database isolated, for example by:

* Using a temporary directory for a test database file.
* Or using an in memory SQLite database configured in `cmdbox.db` for tests.

---

## 4. Summary

The key points of this structure:

* `cmdbox` is the core package name.
* SQLite and Peewee live in `db.py` and `models.py`, optionally wrapped by `repositories.py`.
* The resolver is pure logic that works with model instances but does not control the database.
* The CLI is in `cmdbox.cli.*` and exposes the `cb` command.
* Execution is separated in `cmdbox.execution.executor`.

This layout should keep the codebase clean and make it easy to evolve CmdBox without large refactors.

```
