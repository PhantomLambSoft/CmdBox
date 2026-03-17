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

The fields of a command that you can edit are:
- `alias` - The name you use to recall your command.
- `template` - The part of the command that actually runs in the shell.
- `description` - A description of what the command does.
- `tags` - A list of tags that you can use to categorize your commands.

Each stored command also has additional metadata fields that are stored/updated by CmdBox, and are not editable. These 
are meant to provide you with extra information about the command that you may find useful. They are:
- `date_created`
- `last_updated`
- `used`
- `last_used`
