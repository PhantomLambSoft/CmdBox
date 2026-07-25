# Changelog

## [Unreleased]

### Fixed

- Migration runner not migrating the database correctly.


## [1.2.0] - 2026-07-16

### Added
- Unrecognized top-level commands now default to the `cmd` subcommand group, so `cb add ...` now works the same as
 `cb cmd add ...`. The `cmd` prefix is still supported also.
- Paging to list and search commands for cmd, var, and tag.
- New user defined defaults available in settings
    - Limit for list and search
    - Default order for list and search
    - Pager functionality defaults

### Fixed
- Duplicate shorthand flags on various commands
- 2-letter shorthand flags that did not work
- Unable to set secondary cmd params to null
  - `cwd`, `env`, `shell`, and `timeout`, could not be set back to null once they were set. Each param now has a
  `--clear-*` flag to reset them to null.
  - These params are no longer prompted for in cmd update's `--edit` mode.
- `cb settings data-dir` not working on platforms other than Windows.


## [1.1.3] - 2026-07-08

### Fixed
- Alias commands (`cb <alias>`, `cb tags`, etc.) crashed on fresh install.

### Changed
- Removed the `click` dependencey added in the yanked 1.1.2 release.


## [1.1.2] - 2026-07-08 [YANKED]

### Fixed
- Added missing `click` dependency to pyproject.toml


## [1.1.1] - 2026-07-01

### Fixed
- Prompt for variable when no stored value is found (previously raised NameNotFound error instead)


## [1.1.0] - 2026-06-28

### Added
- Added import/export functionality
    - Commands, variables, and tags are exported to and imported from JSON files.
    - Filter exports by tag, alias/name, and type.
    - Export nested commands and variables by default, with a `--flatten` flag to inline all references.
    - Import validates for circular references and rejects the file before writing anything.
    - Import supports a `--preview` flag to see what would be imported without making any changes.
    - Import supports an `--overwrite` flag to replace existing items on conflict.
- Added the `data-dir` subcommand to settings.
    - Allows users to quickly open CmdBox's data directory in the default file manager.

### Fixed
- Vendor name changed to "PhantomLamb"
- Connected missing settings values to app functionality


## [1.0.0] - 2026-06-11

### Added
- Initial public release