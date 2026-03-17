# cmd

The `cmd` command is the core of CmdBox. Use it to add, remove, edit, list, and inspect your saved commands.

The available subcommands for the cmd module are:
- [add](#add)
- [get](#get)
- [edit](#edit)
- [list](#list)
- [search](#search)
- [delete](#delete)
- [tag](#tag)
- [untag](#untag)

Some of these subcommands also have aliases available. These will be discussed in the subcommands section.

## Add

The `add` subcommand is how you add new commands to your CmdBox database.

When creating a command, you can provide an alias, template, and description in one line. Alias and template are arguments,

```console
> cb cmd add my-command "echo Hello World!" --description "This is a command that prints 'Hello World!' to the console."
```

Alias and template do not a leading name. Description does.

The field that can be entered when creating a command are:
- Alias
- Template
- Description
- Tags

Any of these fields can be left blank, and you will be prompted for them. If you do not wish to provide them, you can 
simply leave them blank, and press enter.

```console
> cb add
```

