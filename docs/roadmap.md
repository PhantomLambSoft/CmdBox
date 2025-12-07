# CmdBox Roadmap

This document describes the near term roadmap for CmdBox with a focus on shipping a solid MVP quickly.  
Items are grouped into phases. Each phase has clear goals and definitions of done.

Naming:

- Product: CmdBox
- Package: `cmdbox`
- CLI binary: `cb`

---

## Phase 1: Core design and specifications

Goal: Lock in the core behavior and public contract so future work aligns with a stable design.

Tasks:

1. Finalize core behavior spec
   - `cmdbox_design.md` with:
     - Command and variable definitions.
     - Interpolation syntax.
     - Resolution order.
     - Error conditions and edge cases.

2. Finalize CLI contract
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

## Phase 2: Core engine implementation

Goal: Implement the pure resolver and models with full test coverage and no dependency on I/O.

Tasks:

1. Implement models
   - `cmdbox.models.Command`
   - `cmdbox.models.Variable`
   - Keep them simple dataclasses.

2. Implement error types
   - `cmdbox.errors` with all known error classes.

3. Implement resolver
   - `cmdbox.resolver` functions for:
     - Resolving a string given commands and variables.
     - Resolving a command by alias.
   - Token parsing and expansion.
   - Cycle detection.
   - Undefined reference handling.

4. Core tests
   - `tests/test_resolver_core.py` mirroring all examples from the design doc.
   - Include success paths and error paths.

Definition of done:

- A Python script inside the repo can import `cmdbox.resolver` and resolve templates directly with in memory models.
- All resolution rules in the design doc pass their tests.

---

## Phase 3: Storage layer MVP

Goal: Add a simple file based backend behind repository interfaces, without changing core logic.

Tasks:

1. Define repository interfaces
   - `cmdbox.storage.base.CommandRepository`
   - `cmdbox.storage.base.VariableRepository`

2. Implement JSON storage backend
   - `cmdbox.storage.json_store`:
     - Load commands and variables from JSON.
     - Save and delete operations.
     - Handle missing or empty files gracefully.

3. Implement path helpers
   - `cmdbox.storage.paths`:
     - Default data directory selection.
     - Functions to compute file paths.
     - Optional override via environment variable or config.

4. Storage tests
   - `tests/test_storage_json_store.py`:
     - Round trip persistence.
     - Behavior on first run (no files yet).
     - Basic error handling.

Definition of done:

- A small script can:
  - Create commands and variables.
  - Save them to disk through the repositories.
  - Load them back and pass them to the resolver.

---

## Phase 4: CLI and execution layer

Goal: Implement `cb` as a thin layer over core and storage, wired to an executor for actual command execution.

Tasks:

1. Choose a CLI framework and set up `cb`
   - Implement `cmdbox.cli.main` as the top level entry point.
   - Register subcommands as defined in `cb_cli_contract.md`.

2. Implement command management subcommands
   - `cb list`
   - `cb show`
   - `cb add`
   - `cb edit`
   - `cb remove`

3. Implement variable management subcommands
   - `cb vars list`
   - `cb vars show`
   - `cb vars add`
   - `cb vars edit`
   - `cb vars remove`

4. Implement run subcommand
   - `cb run <alias> [options]`
   - Wire to resolver and executor.
   - Support `--dry-run` and basic error handling.

5. Execution abstraction
   - `cmdbox.execution.executor.Executor`:
     - Minimal implementation using `subprocess.run`.
   - Potential simple dry run mode in CLI.

6. CLI tests
   - `tests/test_cli/`:
     - Command CRUD flows.
     - Variable CRUD flows.
     - Running commands including error paths.
     - Exit code behavior.

Definition of done:

- Installing the project in a virtual environment exposes `cb` on the path.
- A user can:
  - Add a variable.
  - Add a command.
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
     - Storage errors.
   - Helpful suggestions where appropriate.

2. Help text and documentation
   - Descriptive `--help` content for all subcommands.
   - README quick start:
     - Install.
     - Define variables.
     - Define commands.
     - Run commands.
   - Examples of nested commands and nested variables.

3. Configuration
   - Optional config file support for:
     - Custom data directory.
   - Flags for overriding config per call.

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

3. Optional quality of life improvements
   - Aliases for common CLI commands if helpful.
   - Shortcuts or convenience flags where they do not conflict with the CLI contract.

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

2. Sync and backup
   - Export and import.
   - Remote backups.

3. Team features
   - Shared command libraries.
   - Read only vs local overrides.

4. Advanced ergonomics
   - TUI or simple GUI on top of the same core.
   - Fuzzy search or interactive command picker.

These items should be designed so they can be layered on without breaking existing usage.

---

## Summary

The roadmap for CmdBox focuses first on:

1. Stable design and contracts.
2. A pure and well tested core.
3. A small but solid file backed storage system.
4. A clean `cb` CLI that developers can rely on.

Once the MVP is stable and pleasant to use, later phases extend functionality and ergonomics while preserving the existing mental model and CLI surface.
