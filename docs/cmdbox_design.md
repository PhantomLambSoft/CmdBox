# CmdBox MVP Design

This document defines the core behavior, data structures, and resolution rules for CmdBox.
It is the foundation for the core engine and ensures the rest of the system stays consistent as the app grows.

---

## 1. MVP Scope

CmdBox is a local command library that lets users:

1. Define commands with an alias and a template string.
2. Define variables that can be used in commands or other variables.
3. Nest commands inside other commands.
4. Resolve a command by replacing all variable references and nested command references.
5. Output or execute the final resolved command in the users shell.

Out of scope for MVP:

* Multi user support.
* Cloud sync.
* Profiles or environments.
* Shell specific init scripts.
* GUI or TUI.
* Permissions or credential systems.

---

## 2. Core Concepts

### 2.1 Command

A command is a user defined template with an alias.

Fields:

* alias: string
* template: string
* description: optional string
* tags: optional list of strings

Example:

```
alias: go-home
template: cd <home_path>
description: Jump to the main work directory.
```

Another example using nested variables:

```
alias: serve
template: cd <project_root> && npm start
```

Example using nested command references:

```
alias: deploy
template: <cmd:build> && rsync <deploy_path> user@server:/var/www/html
```

---

### 2.2 Variable

A variable is a simple key value pair.

Fields:

* name: string
* value: string

A variable may refer to other variables using the same placeholder syntax.

Example:

```
name: home_path
value: C:/Users/Max/PycharmProjects/cmdbox
```

Nested variable example:

```
name: project_root
value: <home_path>/projects/app
```

---

### 2.3 Interpolation Syntax

CmdBox uses angle bracket tokens for references.

* Variable reference: `<variable_name>`
* Command reference: `<cmd:alias>`

These tokens must be replaced at resolution time.

Examples:

```
cd <home_path>
```

```
docker run -v <home_path>:/data <cmd:docker-image>
```

Token rules:

* Names are alphanumeric plus underscore.
* No whitespace inside brackets.
* Anything that does not match `<name>` or `<cmd:name>` is preserved as literal text.

---

## 3. Resolution Rules

The resolution process takes one command alias and produces a fully expanded string.

The resolver follows these steps:

### 3.1 Load the command

If the alias does not exist, resolution fails with UnknownAlias.

### 3.2 Replace command references

Any `<cmd:alias>` token is replaced by the resolved template of that command.
This requires recursive resolution.

Example:

```
build = npm run build
deploy = <cmd:build> && rsync ./dist ...
```

### 3.3 Replace variable references

Variable tokens `<name>` are replaced by the variables value.
This may trigger nested resolution if the value contains more tokens.

Example:

```
project_root = <home_path>/projects/app
```

### 3.4 Order of expansion

Commands inside commands are resolved first.
Variables are resolved after command expansion.

Justification:

* Commands are explicit logical units.
* Variable interpolation is pure textual substitution.

### 3.5 Cycle detection

The resolver must detect cycles in both variables and commands.

Examples:

Variable cycle:

```
A = <B>
B = <A>
```

Command cycle:

```
build = <cmd:deploy>
deploy = <cmd:build>
```

Both must raise a ResolutionError.

### 3.6 Undefined references

If a variable or nested command reference is encountered that does not exist, resolution fails with:

* UnknownVariable
* UnknownCommandReference

There is no automatic fallback.

### 3.7 Literal escaping (MVP choice)

MVP behavior:

* There is no escape syntax for angle brackets.
* If a user wants literal angle brackets, they must use a workaround like doubling them `<<` or a variable.

Escaping rules can be added later without breaking existing commands.

---

## 4. Examples

### 4.1 Basic variable interpolation

Variables:

```
home_path = C:/Users/Max
```

Command:

```
alias: go-home
template: cd <home_path>
```

Resolved output:

```
cd C:/Users/Max
```

---

### 4.2 Nested variables

Variables:

```
home_path = C:/Users/Max
project_root = <home_path>/workspace/cmd
```

Command:

```
alias: open-project
template: cd <project_root>
```

Resolved output:

```
cd C:/Users/Max/workspace/cmd
```

---

### 4.3 Commands inside commands

Commands:

```
alias: build
template: npm run build

alias: deploy
template: <cmd:build> && scp -r ./dist user@server:/var/www/html
```

Resolved output:

```
npm run build && scp -r ./dist user@server:/var/www/html
```

---

### 4.4 Commands using variables that call commands

Variables:

```
build_output = dist
```

Commands:

```
alias: build
template: npm run build

alias: deploy
template: <cmd:build> && rsync <build_output> server:/var/www/html
```

Resolved output:

```
npm run build && rsync dist server:/var/www/html
```

---

### 4.5 Error case: missing variable

Command:

```
template: cd <not_defined>
```

Resolution should fail with UnknownVariable.

---

### 4.6 Error case: command cycle

```
a = <cmd:b>
b = <cmd:a>
```

Resolution should fail with ResolutionError.

---

## 5. Edge Cases and Clarifications

### 5.1 Empty variable value

If a variable exists but is empty:

```
temp = 
```

Then `<temp>` resolves to an empty string.
This does not cause an error.

### 5.2 Unused variables

Variables that are never referenced are allowed.

### 5.3 Duplicate alias or variable name

Attempting to define an alias or variable that already exists should:

* overwrite (simple option), or
* disallow overwrites unless forced.

For MVP, allow overwrites with a warning.

### 5.4 Whitespace handling

CmdBox does not trim whitespace automatically.
The template is taken exactly as written.

### 5.5 Parameterized commands (future feature)

MVP does not support passing parameters into commands.
All templates are static except for variable references.
This keeps the first version simple and predictable.

---

## 6. Data Model Summary

### Command

```
Command {
  alias: str
  template: str
  description: str optional
  tags: list[str] optional
}
```

### Variable

```
Variable {
  name: str
  value: str
}
```

### Resolver Behavior Summary

* resolve(alias) returns a string or raises an error.
* recursively resolve command references.
* recursively resolve variable references.
* detect and fail on cycles.
* fail on undefined references.

---

## 7. Minimal Expected Capabilities for MVP

CmdBox MVP must be capable of:

1. Storing user defined commands.
2. Storing user defined variables.
3. Resolving a command template into a final string.
4. Running or printing that string.
5. Handling nested references.
6. Handling cycle detection.
7. Handling missing references gracefully.
8. Working fully offline.
