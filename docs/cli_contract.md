# CmdBox CLI Contract

This document defines the public command line interface for the CmdBox MVP.
The goal is to establish a stable surface so future changes to the core engine or storage layer do not break user
workflows.

The CLI is organized into two main groups:

1. Commands for managing and executing command templates.
2. Commands for managing variables.

Everything here is intended to remain backward compatible after release of the MVP.

---

## 1. Top Level Command

The application entry point is:

```
cb <subcommand> [options]
```

The CLI uses subcommands with positional arguments and optional flags.
The CLI should provide clear help text at every level:

```
cb --help
cb add --help
cb vars --help
```

---

# 2. Command Management

These commands manage user defined command templates.

---

## 2.1 List Commands

```
cb list
```

Options:

```
--filter <text>     Filter commands by substring match in alias or description.
--tags <tag,tag>    Only show commands that contain all listed tags.
--json              Output results as JSON.
```

Behavior:

* Prints a table of commands.
* If the user has none, prints a friendly message or an empty table.
* Never throws an error for empty results.

Example:

```
cb list --filter deploy
```

---

## 2.2 Show Command

```
cb show <alias>
```

Behavior:

* Prints all fields of the command:

    * alias
    * template
    * description
    * tags
* Fails with UnknownAlias if not found.

Example:

```
cb show deploy
```

---

## 2.3 Add Command

```
cb add <alias> --template "<template>" [options]
```

Options:

```
--template "<text>"     Required. The command template.
--description "<text>"  Optional. Description text.
--tags "tag1,tag2,tag3" Optional. Comma separated tags.
--force                 Overwrite existing command if alias already exists.
```

Behavior:

* Saves a new command.
* If alias exists and --force is not specified: fail with AliasExists.

Example:

```
cb add go-home --template "cd <home_path>" --description "Jump to home"
```

---

## 2.4 Edit Command

```
cb edit <alias> [options]
```

Options:

```
--template "<new template>"
--description "<new description>"
--tags "tag1,tag2"
```

Behavior:

* Any provided fields overwrite existing values.
* If alias does not exist: fail with UnknownAlias.
* If nothing is provided: show help.

Example:

```
cb edit deploy --template "<cmd:build> && rsync dist server:/var/www/html"
```

---

## 2.5 Remove Command

```
cb remove <alias>
```

Options:

```
--yes    Skip confirmation.
```

Behavior:

* Deletes the command.
* If alias does not exist: fail with UnknownAlias.

Example:

```
cb remove serve
```

---

## 2.6 Run Command

```
cb run <alias> [<args>...]
```

Options:

```
--dry-run      Resolve the command but do not execute it.
--no-print     Do not print the resolved command before executing.
--json         Output resolved command as JSON.
```

Behavior:

1. Loads user commands and variables.
2. Resolves the target alias.
3. If errors occur:

    * UnknownAlias
    * UnknownVariable
    * UnknownCommandReference
    * ResolutionError (cycle or other failure)
4. If dry-run:

    * Output the fully resolved command string.
    * Do not execute.
5. If not dry-run:

    * Print the resolved command unless suppressed.
    * Execute through the Executor abstraction.

Examples:

```
cb run deploy
cb run deploy --dry-run
```

Argument passthrough (future enhancement) is intentionally not part of the MVP.

---

# 3. Variable Management

A separate namespace under `cb vars`.

---

## 3.1 List Variables

```
cb vars list
```

Options:

```
--json         Output as JSON.
--filter <text>
```

Behavior:

* Shows all variables.
* Never errors for empty lists.

Example:

```
cb vars list
```

---

## 3.2 Show Variable

```
cb vars show <name>
```

Behavior:

* Prints the variable name and value.
* Fails with UnknownVariable if not found.

Example:

```
cb vars show home_path
```

---

## 3.3 Add Variable

```
cb vars add <name> --value "<value>"
```

Options:

```
--value "<text>"     Required.
--force              Overwrite if the name already exists.
```

Behavior:

* Adds a new variable.
* If variable already exists and --force is not provided: fail with VariableExists.

Example:

```
cb vars add home_path --value "C:/Users/Max"
```

---

## 3.4 Edit Variable

```
cb vars edit <name> --value "<new value>"
```

Behavior:

* Overwrites the value of an existing variable.
* Fails with UnknownVariable if name does not exist.

Example:

```
cb vars edit project_root --value "<home_path>/projects/cmd"
```

---

## 3.5 Remove Variable

```
cb vars remove <name>
```

Options:

```
--yes
```

Behavior:

* Deletes a variable.
* Fails with UnknownVariable if not found.

Example:

```
cb vars remove build_output
```

---

# 4. Global Flags

These apply to all subcommands.

```
--help            Print help text.
--version         Show version.
--config <path>   Use an alternate config file.
--store <path>    Override the storage directory for this invocation.
--json            Output results in JSON when applicable.
```

---

# 5. Exit Codes

Standardized exit codes for predictable scripting.

* 0: success
* 1: generic error
* 2: unknown alias
* 3: unknown variable
* 4: resolution error (cycle, missing nested reference)
* 5: storage error (cannot read or write)
* 6: CLI usage error (invalid flags, missing arguments)

---

# 6. Examples of Complete Flows

## Add a variable, add a command, run it

```
cb vars add home_path --value "C:/Users/Max"
cb add go-home --template "cd <home_path>"
cb run go-home
```

Output:

```
cd C:/Users/Max
```

---

## Edit a command, then dry run

```
cb edit deploy --template "<cmd:build> && rsync dist server:/var/www/html"
cb run deploy --dry-run
```

Output:

```
npm run build && rsync dist server:/var/www/html
```

---

## Delete a variable

```
cb vars remove home_path --yes
```

---

# 7. CLI Stability Guarantees

These rules ensure the CLI remains stable across updates.

1. Existing subcommands will not be removed.
2. Flag names will not change after MVP release.
3. The meaning of placeholders `<name>` and `<cmd:name>` will not change.
4. New features must be introduced as additional flags or new subcommands.
5. Breaking changes will only occur with a major version bump.
