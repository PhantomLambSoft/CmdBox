# cmd

The `cmd` subcommand is the core of CmdBox. Use it to add, remove, edit, list, and inspect your saved commands.

The available subcommands for the `cmd` module are:

- [add](#add)
- [get](#get)
- [edit](#edit)
- [list](#list)
- [search](#search)
- [delete](#delete)
- [tag](#tag)
- [untag](#untag)

Some of these subcommands also have aliases available. These will be discussed in the subcommands section.

## `add`

The `add` subcommand adds new commands to your CmdBox database.

When creating a command, you only have to provide an alias. You will be prompted for the rest of the fields.

```console
$ cb cmd add prune-docker

? Enter template: docker system prune -f
? Enter description: Removes all stopped containers.
? Enter tags (comma-separated): dev,docker
```

This same command can be entered in one line.

```console
$ cb cmd add prune-docker "docker system prune -f" --description "Removes all stopped containers."
```

Notice in this example that alias and template are provided without a flag. They are not optional fields. Description
is an optional field, so it must be prefaced with a `--description` flag.

You will always be prompted for tags.

!!! tip "Autocomplete"
    Autocomplete is available for stored tags. In the tag prompt, start typing and the available tags will be suggested.

If you want to be prompted for every input, you can use the `--interactive` (or `-i`) flag.

```console
$ cb cmd add --interactive

? Enter alias: prune-docker
? Enter template: docker system prune -f
? Enter description: Removes all stopped containers.
? Enter tags (comma-separated): dev,docker
```

!!! tip
    Template prompts are multi-line by default. To submit a single-line template, press **Escape** and then **Enter** 
    after typing your template text.

### Execution context

You can optionally store execution context on a command. These values are used as defaults each time the command runs
and can be overridden at runtime using the same flags on `cb run`.

| Option | Description |
|---|---|
| `--cwd` | Working directory to run the command from |
| `--shell` | Shell to use when running the command |
| `--env KEY=VALUE` | Environment variable to set (repeatable) |
| `--timeout` | Maximum number of seconds before the process is killed |

```console
$ cb cmd add deploy "npm run deploy" --cwd "/home/user/projects/myapp" --env NODE_ENV=production --timeout 60
```

The `--env` flag can be supplied multiple times to set more than one variable.

```console
$ cb cmd add api-start "python main.py" --env HOST=0.0.0.0 --env PORT=8080 --env DEBUG=false
```

!!! note "Windows shells and environment variables"
    On Windows, templates using `%VAR%` syntax require `--shell cmd.exe` to be stored on the command. Without it,
    the variable reference will be treated as a literal string rather than expanded.

```console
$ cb cmd add show-path "echo %PATH%" --shell cmd.exe
```

## `get`

The `get` command retrieves a command and displays all of its available fields along with its tags.

```console
$ cb cmd get upgrade-pip
```

!!! note
    All outputs are stylized. Some of the more stylized outputs will be displayed here in a different format, as 
    shown below:

![Command get output](../assets/cmd/cmd-get-output.svg)

When execution context has been stored on a command, it will appear in the output alongside the other fields. Otherwise,
it will not be shown.

## `update`

!!! tip "Aliases"
    `update` can also be called as `edit`.

The `update` command is used to make changes to a command you already have stored.

You can change only a specific field by specifying that field along with the new value you want it to have.

```console
$ cb cmd update prune-docker --description "Removes all stopped containers, dangling images, and unused networks."
```

!!! warning
    Be sure to wrap your values in quotes if they contain spaces.

This updates only the description of the stored `prune-docker` command.

Multiple fields can be updated at once by using the `--set` flag and using key value pairs like `key=value`. Each pair
should be separated by a comma with no spaces.

```console
$ cb cmd update prune-docker --set template="docker system prune",description="Removes all stopped containers, asking for confirmation"
```

!!! tip "Autocomplete"
    Autocomplete is available for available fields when using `--set`.

### Updating execution context

The execution context fields can be updated using the same flags as `add`.

```console
$ cb cmd update deploy --cwd "/home/user/projects/myapp" --timeout 120
```

The `--env` flag replaces the stored environment variables with the new values you supply. To set multiple variables,
supply the flag multiple times.

```console
$ cb cmd update api-start --env HOST=0.0.0.0 --env PORT=9090
```

### Edit mode

If you want to update the current value of any field without supplying a completely new value, you can use the `--edit`
(or `-e`) flag. When using `--edit`, you will be prompted to update each field, and the prompt will be pre-filled with
the current value.

```console
$ cb cmd update prune-docker --edit

? Enter alias: prune-docker
? Enter template: docker system prune
? Enter description: Removes all stopped containers, dangling images, and unused networks.
```

If you know you only want to update a specific field and you do not want to iterate through each field, you can specify
which fields to update by using the `--edit-fields` flag.

```console
$ cb cmd update prune-docker --edit --edit-fields description

? Enter description: Removes all stopped containers, dangling images, and unused networks.
```

The `--edit-fields` flag also supports the execution context fields `cwd`, `shell`, and `timeout`.

```console
$ cb cmd update deploy --edit --edit-fields cwd

? Working directory: /home/user/projects/myapp
```

!!! note
    Environment variables (`env`) cannot be updated in edit mode. Use the `--env` flag directly instead.

!!! warning
    `--edit-fields` can only be used in conjunction with the `--edit` flag.

## `list`

!!! tip "Aliases"
    `list` can also be called as `ls`.

The `list` command displays all commands you have stored in your database.

```console
$ cb cmd list
```

![Command list output](../assets/cmd/cmd-list-output.svg)

By default, only the alias, template, and description of each command are displayed, and the default order is by alias.
The default fields and ordering can be adjusted in your settings, or by supplying additional options to the `list`
command.

To change the order, use the `--order` flag and specify the field you want to order by.

```console
$ cb cmd list --order description
```

![Command list order by description output](../assets/cmd/cmd-list-order-by-description-output.svg)

To change the displayed fields, use the `--field` flag and specify the fields you want to display.

```console
$ cb cmd list --field alias --field template
```

![Command list alias template only output](../assets/cmd/cmd-list-alias-template-only-output.svg)

If you have a large database of commands, you may only want to list some of them. For this, you can use the `--limit`
flag.

```console
$ cb cmd list --limit 10
```

List can also be limited to only commands that feature a specific tag.

```console
$ cb cmd list --tag dev
```

The `--tag` flag can be used multiple times to list commands that feature multiple tags.

```console
$ cb cmd list --tag dev --tag docker --tag production
```

!!! tip
    When using multiple `--tag` flags, commands that feature any of those tags will be displayed.

## `search`

!!! tip "Aliases"
    `search` can also be called as `find`.

While `list` lets you filter your commands by tag, search lets you filter your commands by any of the available fields.
By default, search is limited to the alias, template, and description fields.

```console
$ cb cmd search pip
```

![Command search output](../assets/cmd/cmd-search-output.svg)

Using the `--in` flag, you can limit your search to only the fields you want to search in.

```console
$ cb cmd search listening --in description
```

![Command search in description output](../assets/cmd/cmd-search-in-description-output.svg)

And if you only want to see certain fields in the results output, you can use the `--field` flag.

```console
$ cb cmd search listening --in description --field alias
```

![Command search field alias output](../assets/cmd/cmd-search-field-alias-output.svg)

As with the list command, you can also limit the number of results returned by using the `--limit` flag.

```console
$ cb cmd search pip --limit 3
```

## `delete`

!!! tip "Aliases"
    `delete` can also be called as `rm`, `del`, or `remove`.

The `delete` command is used to remove a command from your database. It only takes the alias of the command you want
to remove.

```console
$ cb cmd delete prune-docker
```

## `tag`

The `tag` command is used to add tags to a command.

To add a tag, supply the command with the alias of the command you want to tag, followed by the name of the tag.

```console
$ cb cmd tag pip-outdated dev
```

!!! tip "Autocomplete"
    Autocomplete is available for tags.

## `untag`

Untag works the same as `tag`.

```console
$ cb cmd untag pip-outdated dev
```
