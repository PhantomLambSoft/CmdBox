```
cb/
├─ pyproject.toml
├─ README.md
├─ LICENSE
├─ .gitignore
├─ docs/
│ ├─ cmdbox_design.md # Phase 1 spec (core concepts and resolution rules)
│ ├─ cb_cli_contract.md # CLI contract doc (using cb as the command)
│ ├─ architecture_overview.md # High level overview and future notes
│ └─ roadmap.md # High level milestones and ideas
├─ src/
│ └─ cb/
│ ├─ __init__.py
│ ├─ __main__.py # Entry point for `python -m cb` (delegates to CLI)
│ ├─ config.py # Config loading and defaults
│ ├─ models.py # Command and Variable data models
│ ├─ errors.py # Error types (UnknownAlias, UnknownVariable, etc.)
│ ├─ resolver.py # Pure core engine for interpolation and resolution
│ ├─ storage/
│ │ ├─ __init__.py
│ │ ├─ base.py # Repository interfaces and abstract base classes
│ │ ├─ json_store.py # MVP JSON file based storage implementation
│ │ └─ paths.py # Functions for resolving storage paths on disk
│ ├─ execution/
│ │ ├─ __init__.py
│ │ └─ executor.py # Executor abstraction for running resolved commands
│ ├─ cli/
│ │ ├─ __init__.py
│ │ ├─ main.py # Top level CLI handler used by console_script "cb"
│ │ ├─ commands.py # Command related subcommands (list, add, show, run, etc.)
│ │ └─ vars.py # Variable related subcommands (list, add, edit, etc.)
│ └─ logging_setup.py # Optional simple logging setup for debug flags
└─ tests/
├─ __init__.py
├─ test_models.py
├─ test_resolver_core.py # Pure engine tests
├─ test_storage_json_store.py # JSON backend tests
├─ test_execution_executor.py # Execution abstraction tests (can be light at first)
└─ test_cli/
├─ __init__.py
├─ test_cli_commands.py # Tests for `cb add`, `cb list`, `cb show`, `cb run`
└─ test_cli_vars.py # Tests for `cb vars list`, `cb vars add`, etc.
```
