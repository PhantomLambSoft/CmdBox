# CmdBox Architecture Overview

This document describes the high level architecture for CmdBox.

- Product name: CmdBox
- Python package name: `cmdbox`
- CLI binary name: `cb`

CmdBox is a local command library and runner. Users define reusable command templates and variables, then use `cb` to
resolve and execute them in their shell.

The architecture is split into three main layers:

1. Core engine
2. Storage layer
3. CLI and execution layer

Each layer has a clear responsibility and communicates with the others through narrow interfaces. This keeps refactoring
low risk and makes future features mostly additive.

---

## 1. Layered Architecture

### 1.1 Core Engine

The core engine is pure logic and has no side effects.

Responsibilities:

- Represent commands and variables.
- Resolve templates by expanding variable references and nested command references.
- Detect and report errors such as cycles and missing references.

Key modules:

- `cmdbox.models`
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
        - Accept collections of `Command` and `Variable` instances.
        - Perform interpolation of `<name>` and `<cmd:alias>` tokens.
        - Detect cycles and undefined references.
        - Return a resolved string or raise a well defined error.

Design principles:

- No access to the file system.
- No subprocess calls.
- No printing or logging to the terminal.
- Pure functions where possible so they are easy to test and reason about.

---

### 1.2 Storage Layer

The storage layer is responsible for persistence of commands and variables.  
It wraps all file system interaction behind repository interfaces.

Responsibilities:

- Load and save `Command` and `Variable` instances.
- Hide storage details from the rest of the system.
- Provide a stable interface so the storage backend can be changed without touching the core or CLI.

Key modules:

- `cmdbox.storage.base`
    - `CommandRepository` abstract base class
    - `VariableRepository` abstract base class

- `cmdbox.storage.json_store`
    - MVP implementation that stores data in JSON files on disk.
    - Responsible for serialization and deserialization.

- `cmdbox.storage.paths`
    - Functions for determining where to store data on disk:
        - Default data directory
        - Commands file path
        - Variables file path
    - Honors configuration overrides where applicable.

Design principles:

- Repository interfaces are small and focused:
    - `list`
    - `get`
    - `save`
    - `delete`
- Core engine depends only on the repository interfaces, not on the JSON implementation.
- Storage errors are translated into `CbError` subclasses or wrapped cleanly.

---

### 1.3 CLI and Execution Layer

The CLI and execution layer is the outermost shell facing part of CmdBox.  
It connects user input to the core engine and storage.

Responsibilities:

- Parse command line arguments.
- Call into repositories and the resolver.
- Format and print outputs for the user.
- Execute resolved commands in the environment if requested.
