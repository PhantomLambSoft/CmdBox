# CmdBox Architecture Overview

This document describes the high level architecture for CmdBox.

- Product name: CmdBox
- Python package name: `cmdbox`
- CLI binary name: `cb`

CmdBox is a local command library and runner. Users define reusable command templates and variables, then use `cb` to
resolve and execute them in their shell.

The architecture is split into three main layers:

1. Core engine
2. Data access layer (SQLite with Peewee)
3. CLI and execution layer

Each layer has a clear responsibility and communicates with the others through narrow interfaces. This keeps refactoring
low risk and makes future features mostly additive.

---

## 1. Layered Architecture

### 1.1 Core Engine

The core engine is pure logic and has no side effects.

Responsibilities:

- Define the conceptual behavior of commands and variables.
- Resolve templates by expanding variable references and nested command references.
- Detect and report errors such as cycles and missing references.

Core engine code should treat Peewee models as simple in memory objects. It should not open connections, run queries, or
know about database configuration.

Key modules:

- `cmdbox.models`
    - Peewee `Model` subclasses:
        - `Command`
        - `Variable`

- `cmdbox.errors`
    - `CbError` base class
    - `UnknownAlias`
    - `UnknownVariable`
    - `ResolutionError`
    - `UnknownCommandReference`
    - Other error types as needed.

- `cmdbox.resolver`
    - Core resolution functions that:
        - Accept `Command` and `Variable` objects passed in from the data access layer.
        - Perform interpolation of `<name>` and `<cmd:alias>` tokens.
        - Detect cycles and undefined references.
        - Return a resolved string or raise a well defined error.

Design principles:

- No direct access to the database connection from resolver code.
- No subprocess calls.
- No printing or logging to the terminal.
- Logic is deterministic and easy to unit test using in memory instances or a temporary SQLite database.

---

### 1.2 Data Access Layer (SQLite with Peewee)

The data access layer is responsible for persistence of commands and variables.  
It uses SQLite as the database engine and Peewee as the ORM.

Responsibilities:

- Define the database schema for commands and variables.
- Connect to the SQLite database file.
- Run queries to list, get, create, update, and delete records.
- Provide a stable interface so the underlying database can be changed later if needed.

Key modules:

- `cmdbox.db`
    - Database configuration:
        - A Peewee `SqliteDatabase` instance.
        - Initialization helpers for setting up the database file and creating tables.

- `cmdbox.models`
    - Peewee `Model` subclasses:
        - `Command`
        - `Variable`
    - Typical fields:
        - `Command.alias`, `Command.template`, `Command.description`, `Command.tags`
        - `Variable.name`, `Variable.value`

- `cmdbox.repositories`
    - Optional repository layer on top of Peewee models, such as:
        - `CommandRepository`
        - `VariableRepository`
    - These provide methods:
        - `list`
        - `get`
        - `create`
        - `update`
        - `delete`

Design principles:

- The rest of the code should not know about SQL directly, only Peewee models or repository methods.
- Database paths are configurable (for example via config or environment).
- Migrations can be handled either manually or through Peewee helper tools, but initial versions will be simple table
  create operations.

---

### 1.3 CLI and Execution Layer

The CLI and execution layer is the outermost shell facing part of CmdBox.  
It connects user input to the data access layer and core engine.

Responsibilities:

- Parse command line arguments.
- Call into Peewee backed repositories and the resolver.
- Format and print outputs for the user.
- Execute resolved commands in the environment if requested.

Key modules:

- `cmdbox.cli.main`
    - Defines the `cb` CLI application entry point.
    - Registers subcommands such as:
        - `list`
        - `show`
        - `add`
        - `edit`
        - `remove`
        - `run`
        - `vars list`
        - `vars add`
        - `vars edit`
        - `vars remove`

- `cmdbox.cli.commands`
    - Implementation of top level command management CLI handlers.

- `cmdbox.cli.vars`
    - Implementation of variable management CLI handlers.

- `cmdbox.execution.executor`
    - Executes a resolved command string using the system shell.
    - Provides a simple interface such as:
        - `run(command: str) -> int`

- `cmdbox.config`
    - Optional configuration support:
        - SQLite database file path
        - Other behavioral options

Design principles:

- CLI code is as thin as possible.
- Resolver logic and database access logic stay outside of CLI code.
- The CLI wires together:
    - the database connection
    - the models or repositories
    - the resolver
    - the executor

---

## 2. Data Flow

This section describes how data flows through the system for common operations using Peewee and SQLite.

### 2.1 Running a command

User runs:

```bash
cb run deploy
````

Flow:

1. `cb` binary invokes `cmdbox.cli.main`.
2. CLI parses `run deploy`.
3. CLI ensures the SQLite database is initialized via `cmdbox.db`.
4. CLI uses Peewee models or repositories to:

    * Load the `Command` with alias `deploy`.
    * Load all `Command` and `Variable` records needed for resolution.
5. CLI calls the resolver:

    * Passes the target command and the full set of commands and variables.
6. Resolver:

    * Validates that `deploy` exists.
    * Recursively resolves `<cmd:...>` references.
    * Recursively resolves `<name>` variable references.
    * Detects any cycles or missing references.
    * Returns resolved string or raises a `CbError`.
7. CLI:

    * If `--dry-run` is set:

        * Prints resolved command.
        * Does not execute.
    * Else:

        * Optionally prints resolved command.
        * Calls `Executor.run(resolved)` to execute.
        * Returns exit code as the process exit code.

Errors like `UnknownAlias` or `UnknownVariable` are caught at CLI level and shown with clear messages.

---

### 2.2 Adding a command

User runs:

```bash
cb add go-home --template "cd <home_path>"
```

Flow:

1. `cb` invokes `cmdbox.cli.main`.
2. CLI parses the `add` subcommand and its options.
3. CLI ensures `cmdbox.db` is initialized.
4. CLI uses Peewee `Command` model or repository to:

    * Check if alias already exists.
    * Create a new `Command` row in SQLite.

This path does not touch the resolver.

---

### 2.3 Listing commands

User runs:

```bash
cb list
```

Flow:

1. CLI parses `list`.
2. CLI initializes the database.
3. CLI calls:

    * `Command.select()` or a repository `list` method.
4. CLI applies any filtering in Python if needed.
5. CLI prints a table or JSON output as requested.

Again, this path does not touch the resolver or executor.

---

## 3. Separation of Concerns

The architecture is designed so that each major concern lives in a single place.

* Resolution logic and token expansion:

    * `cmdbox.resolver`
* Error types:

    * `cmdbox.errors`
* Database schema and access:

    * `cmdbox.db`
    * `cmdbox.models`
    * `cmdbox.repositories` (if used)
* Execution:

    * `cmdbox.execution.executor`
* CLI user interface:

    * `cmdbox.cli.*`

This separation provides several benefits:

* The core resolution logic is tested through Peewee models but does not control database connections.
* SQLite can be replaced with a different database later with minimal changes if needed.
* CLI behavior can evolve (new flags, new subcommands) while reusing the same core logic and models.
* Other frontends such as a TUI or GUI can reuse the same core engine and data access layer.

---

## 4. Extensibility

The initial architecture is meant to support:

* Additional tables (profiles, history, tags, etc).
* Future migration tools.
* Alternative backends that still use Peewee but different database engines.

Some examples of future extensions and where they would live:

* Profiles or environments:

    * New tables in `cmdbox.models`.
    * New repository logic in `cmdbox.repositories`.
    * Config and flags in `cmdbox.config` and CLI.

* Sync or export:

    * New modules to export or import data from the SQLite database.
    * Additional CLI commands such as `cb export` and `cb import`.

* Shell init helpers:

    * New CLI subcommands such as `cb init bash`.
    * Modules like `cmdbox.cli.init` that generate shell snippets.

---

## 5. Testing Strategy

Tests are organized to follow the same architecture:

* Core resolution tests:

    * `test_resolver_core.py` exercise all resolution rules defined in the design document.
    * Use an in memory SQLite database (for example `":memory:"`) or a temporary file with Peewee models.

* Data access tests:

    * `test_db_models.py` or `test_repositories.py`:

        * Ensure commands and variables are persisted correctly in SQLite.
        * Validate constraints such as unique aliases and variable names.

* Execution tests:

    * `test_execution_executor.py` verify that commands are passed correctly to the system.

* CLI tests:

    * Located in `tests/test_cli/`.
    * Use a CLI test runner to simulate user calls to `cb`.
    * Validate output formatting and exit codes, including DB setup.

This layered test approach helps keep regressions localized and makes refactoring safer.

---

## 6. Summary

CmdBox is structured as:

* A core resolution engine that operates on Peewee model instances.
* A SQLite based data access layer using Peewee.
* A thin CLI layer that acts as the interface for users via the `cb` command.

The name CmdBox appears in documentation, the package is named `cmdbox`, and the CLI entry point is `cb`. This clear
separation of responsibilities should keep the codebase coherent as the project grows.
