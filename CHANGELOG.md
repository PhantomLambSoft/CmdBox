# Changelog

## [Unreleased]

### Added
- Added import/export functionality
    - Commands, variables, and tags are exported to and imported from JSON files.
    - Filter exports by tag, alias/name, and type.
    - Export nested commands and variables by default, with a `--flatten` flag to inline all references.
    - Import validates for circular references and rejects the file before writing anything.
    - Import supports a `--preview` flag to see what would be imported without making any changes.
    - Import supports an `--overwrite` flag to replace existing items on conflict.

### Fixed
- Vendor name changed to "PhantomLamb"

## [1.0.0] - 2026-06-11

### Added
- Initial public release