# CmdBox Roadmap

This document describes the near term roadmap for CmdBox with a focus on shipping a solid MVP quickly.

Naming:

- Product: CmdBox
- Package: `cmdbox`
- CLI binary: `cb`
- Persistence: SQLite using Peewee models

---

## Phase 1: Core design and specifications

Goal: Lock in the core behavior and public contract so future work aligns with a stable design.

Tasks:

1. Core behavior spec
   - `cmdbox_design.md` with:
     - Command and variable definitions.
     - Interpolation syntax.
     - Resolution order.
     - Error conditions and edge cases.

2. CLI contract
   - `cb_cli_contract.md` with:
     - Subcommand names and arguments.
     - Flags and options.
     - Exit codes.
     - Stability guarantees.

3. Architecture overview
   - `architecture_overview.md` that:
     - Describes the three layers.
     - Defines data flow.
     - Captures separation of concerns.

Definition of done:

- Design docs live in `docs/`.
- There are no open questions about how resolution works or how `cb` behaves from a users perspective.

---

## Phase 2: Database and models (SQLite + Peewee)

Goal: Define the SQLite schema and implement Peewee models for commands and variables.

Tasks:

1. Database setup
   - Create `cmdbox.db` module that:
     - Defines a `SqliteDatabase` instance.
     - Knows how to locate or create the database file.
     - Provides an `init_db()` helper to connect and create tables if needed.

2. Peewee models
   - Implement `cmdbox.models` with Peewee `Model` subclasses:
     - `Command`
     - `Variable`
   - Define fields consistent with the design doc:
     - `Command.alias` (unique), `Command.template`, `Command.description`, `Command.tags` (serialized list or separate table if needed).
     - `Variable.name` (unique), `Variable.value`.

3. Basic CRUD helpers
   - Optionally implement a `cmdbox.repositories` module with small classes:
     - `CommandRepository`
     - `VariableRepository`
   - These wrap Peewee operations for clarity and future flexibility.

4. Model tests
   - `tests/test_db_models.py`:
     - Use a temporary SQLite file or in memory database.
     - Verify that create, read, update, delete all work.
     - Enforce uniqueness constraints for alias and variable name.

Definition of done:

- A simple script can call `init_db()`, create a few commands and variables, and read them back using Peewee.
- Schema is stable enough for the MVP.

---

## Phase 3: Core engine implementation (resolver)

Goal: Implement the resolver that works with Peewee based data, but remains logical and testable.

Tasks:

1. Implement error types
   - `cmdbox.errors` with:
     - `CbError`
     - `UnknownAlias`
     - `UnknownVariable`
     - `UnknownCommandReference`
     - `ResolutionError`

2. Implement resolver
   - `cmdbox.resolver` with functions that:
     - Given a target alias, load the required `Command` and `Variable` records from the database or from repositories.
     - Perform token parsing and expansion.
     - Detect cycles.
     - Fail clearly on missing references.

   - Internally the resolver should work on plain Python values (strings, lists, dicts) when possible, treating Peewee models as simple holders of attributes.

3. Core tests
   - `tests/test_resolver_core.py` mirroring all examples from `cmdbox_design.md`:
     - Variable interpolation.
     - Nested variables.
     - Commands calling commands.
     - Error cases for cycles and missing references.

   - Use an in memory or temporary SQLite database for tests so that realistic Peewee models are used.

Definition of done:

- A Python script inside the repo can:
  - Initialize the database.
  - Insert some commands and variables.
  - Call resolver functions to get fully expanded strings for aliases.

---

## Phase 4: CLI and execution layer

Goal: Implement `cb` as a thin layer over Peewee models, repositories, and the resolver, wired to an executor for actual command execution.

Tasks:

1. CLI setup
   - Implement `cmdbox.cli.main` as the top level entry point.
   - Register subcommands as defined in `cb_cli_contract.md`.

2. Command management subcommands
   - `cb list`
   - `cb show`
   - `cb add`
   - `cb edit`
   - `cb remove`
   - These should:
     - Call `init_db()` on startup.
     - Use Peewee or repositories to talk to the database.
     - Format results for display or JSON output.

3. Variable management subcommands
   - `cb vars list`
   - `cb vars show`
   - `cb vars add`
   - `cb vars edit`
   - `cb vars remove`

4. Run subcommand
   - `cb run <alias> [options]`
   - Flow:
     - Initialize DB.
     - Load necessary commands and variables.
     - Call resolver.
     - If `--dry-run`, print the resolved string only.
     - Otherwise call the executor to run the command.

5. Execution abstraction
   - `cmdbox.execution.executor.Executor`:
     - Minimal implementation using `subprocess.run`.
     - Return exit codes to the CLI.

6. CLI tests
   - `tests/test_cli/`:
     - Simulate `cb` commands.
     - Ensure DB initialization happens once per run.
     - Cover both success and error paths.

Definition of done:

- Installing the project exposes `cb` on the path.
- A user can:
  - Add a variable.
  - Add a command that uses it.
  - Run the command and see it act in their shell.

---

## Phase 5: Fit and finish for MVP

Goal: Bring the first public version of CmdBox to a level that feels reliable and pleasant to use.

Tasks:

1. Error messages and UX polish
   - Human friendly messages for:
     - Unknown alias.
     - Unknown variable.
     - Resolution errors.
     - Database errors.
   - Helpful suggestions when possible.

2. Help text and documentation
   - Descriptive `--help` content for all subcommands.
   - README quick start:
     - Install.
     - Initialize DB automatically on first run.
     - Define variables.
     - Define commands.
     - Run commands.

3. Configuration
   - Optional config file support for:
     - SQLite database file path.
   - Flags for overriding configuration per call.

4. Packaging
   - Finalize `pyproject.toml` with metadata and entry points.
   - `cb` console script mapped to `cmdbox.cli.main`.

5. Basic logging
   - Optional debug logging for troubleshooting.
   - Possibly enabled with a flag like `--debug`.

Definition of done:

- CmdBox can be installed by others and used comfortably with just the README.
- No major known issues in core features.
- Version tagged as an initial release (for example `v0.1.0`).

---

## Phase 6: Shell integration and ergonomics

Goal: Make `cb` feel at home in different shells without changing core behavior.

Tasks:

1. Init helpers
   - New subcommands such as:
     - `cb init bash`
     - `cb init zsh`
     - `cb init powershell`
   - Each prints a small function or alias snippet that:
     - Calls `cb run` and evaluates the result in the current shell.

2. Documentation updates
   - Setup instructions for each supported shell.
   - Tips for integrating into existing dotfiles.

3. Quality of life improvements
   - Small shortcuts or convenience flags where they do not conflict with the CLI contract.

Definition of done:

- On a typical shell, a user can:
  - Add the one time snippet printed by `cb init shell`.
  - Type `cb deploy` and have the resolved command run as if it was typed manually.

---

## Phase 7: Post MVP exploration

Goal: Explore additional value on top of a stable core, without committing to specific features too early.

Possible directions (not all required):

1. Profiles or workspaces
   - Named sets of commands and variables.
   - Switchable with a flag or environment variable.
   - Backed by new tables in SQLite.

2. Sync and backup
   - Export and import from the SQLite database.
   - Remote backups.

3. Team features
   - Shared command libraries stored in a central SQLite or server backed store.
   - Read only vs local overrides.

4. Advanced ergonomics
   - TUI or simple GUI on top of the same core.
   - Fuzzy search or interactive command picker.

These items should be designed so they can be layered on without breaking existing usage.

---

## Summary

The roadmap for CmdBox focuses first on:

1. Stable design and contracts.
2. A SQLite schema with Peewee models.
3. A solid resolver that works on those models.
4. A clean `cb` CLI that developers can rely on.

Once the MVP is stable and pleasant to use, later phases extend functionality and ergonomics while preserving the 
existing mental model and CLI surface.
