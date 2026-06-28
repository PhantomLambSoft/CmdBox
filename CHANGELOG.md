# Changelog

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