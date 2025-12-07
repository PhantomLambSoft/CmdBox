# CmdBox Repository Layout

This document describes the intended layout of the CmdBox repository.

- Product: CmdBox
- Python package: `cmdbox`
- CLI binary: `cb`
- Persistence: SQLite using Peewee models

---

## Top level layout

```
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

