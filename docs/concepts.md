# Concepts

Understanding these core ideas makes everything else in CmdBox click.

## Command

The command object is the heart of CmdBox. It holds the information about the commands you will be executing, as well
as the aliases that will be used to execute them. 

A command has several fields, some of which are editable, and some of which are stored/updated by CmdBox.

The fields of a command that you can edit are:

- `alias` - The name you use to recall your command.
- `template` - The part of the command that actually runs in the shell.
- `description` - A description of what the command does.
- `tags` - A list of tags that you can use to categorize your commands.

Some additional fields are metadata fields that are stored/updated by CmdBox, and are not editable. These
are meant to provide you with extra information about the command that you may find useful. They are:

- `date_created` - The datetime the command was first created.
- `last_updated` - The datetime the command was last updated.
- `used` - A count of how many times the command has been executed.
- `last_used` - The datetime the command was last executed.

Other fields may be accessed in subcommands like `list` or `search` to give you more information about the command or
help narrow your search. You can find more information about these subcommands in the [`cmd` subsetion](commands/cmd.md)

### Alias

An alias is the short name you give a saved command. When you type `cb deploy`, CmdBox looks up the alias `deploy` and 
runs whatever command is stored under it.

Aliases should be:
- **Short** - These are what you're going to be typing to recall commands.
- **Descriptive** - `deploy-prod` beats `dp`, especially if you're searching.
- **Lowercase with hyphens** - This is the conventional style. But aliases can have any characters except white space.

### Template

A command template is the actual shell instructions stored under an alias. It can be anything you would normally type 
in a terminal: a single program, a pipeline, a long string of flags, or even a multi-line script.

```console
alias: git-graph-log
template: git log --oneline --graph --decorate --all
```

Multi-line templates are supported and are executed like a script. In fact, behind the scenes, multi-line templates are 
saved to a temp file and executed as a script. See [`cmd` subsetion](commands/cmd.md) for more information about multi-line 
templates.

## Variables

Variables help make the commands you save more versatile and usable. A variable is defined in a command template by putting 
the variable name inside angle brackets `<var-name>`. This will be replaced when a command is executed.

Variables can be supplied at runtime, for dynamic usage where some portion of a long command may change depending on the situation.

Consider the template: 

`ssh <user>@<host> -p <port>`.

This base template can be used with different variables to connect to different servers.

```console
> cb ssh-server --user admin --host 192.168.1.1 --port 22

ssh admin@192.168.1.1 -p 22
```

Variables can also be stored in the database so they don't have to be entered each time, and they can be used in multiple 
commands.

```console
> cb var add user admin
> cb var add host 192.168.1.1
> cb var add port 22
```

Then when we run the same command as above:

```console
> cb ssh-server

ssh admin@192.168.1.1 -p 22
```

Even if a variable is stored in the database, supplying it at runtime will take precedence. Variables can be mixed and matched
this way:

```console
> cb ssh-server --host 10.0.0.5 --port 2222

ssh admin@10.0.0.5 -p 2222
```

Variables have less metadata than commands, but there are two more fields created by CmdBox that can be used to sort 
variables or just display more information about them.

- `date_created` - The datetime the variable was first created.
- `last_updated` - The datetime the variable was last updated.

## Tags

Tags are labels you attach to commands or variables for organization. A command or variable can have multiple tags. Lists
of commands and variables can be filtered by tag, or searched for by tag:

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
