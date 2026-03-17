# Concepts

Understanding these core ideas makes everything else in CmdBox click.

## Aliases

An alias is the short name you give a saved command. When you type `cb deploy`, CmdBox looks up the alias `deploy` and 
runs whatever command is stored under it.

Aliases should be:
- **Short** - These are what you're going to be typing to recall commands.
- **Descriptive** - `deploy-prod` beats `dp`, especially if you're searching.
- **Lowercase with hyphens** - This is the conventional style. But aliases can have any characters except white space.

## Command Template

A command template is the actual shell instructions stored under an alias. It can be anything you would normally type 
in a terminal: a single program, a pipeline, a long string of flags, or even a multi-line script.

```console
alias: git-graph-log
template: git log --oneline --graph --decorate --all
```

## Variables

Variables let you write one command that works in many situations. Define a placeholder using angle brackets `<>` and 
then use the variable name to replace it.

```console
alias: ssh-server
template: ssh <user>@<host> -p <port>
```

You can supply these values at runtime:

```console
cb ssh-server --user admin --host 192.168.1.1 --port 22
```

Or you can store variables in the database using the [`var` subcommand](commands/var.md) so you don't have to type them 
every time.

## Tags

Tags are labels you attach to commands or variables for organization. A command or variable can have multiple tags.

```console
cb cmd add deploy "..." --tags work,aws,production
```

You can then filter your command list by tag:

```console
cb cmd list --tag aws
```

## Settings

CmdBox behavior and appearance has a lot of customization available through settings. Use the 
[`settings` subcommand](commands/settings.md) to view and change them.
