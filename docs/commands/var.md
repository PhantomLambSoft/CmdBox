# var

The `var` subcommand lets you store variable values in the database to be accessed inside your command templates whenever 
you run a stored command.

While variables stored in command templates can be supplied at runtime, the `var` subcommand deals exclusively with variables 
stored in the database.

The available subcommands for `var` are:

- [`add`](#add)
- [`get`](#get)
- [`update`](#update)
- [`list`](#list)
- [`search`](#search)
- [`delete`](#delete)
- [`tag`](#tag)
- [`untag`](#untag)

Some of these subcommands also have aliases available. These will be discussed in the subcommands section.

## `add`

The add command adds a variable to the database.

When creating a variable, you only have to supply the name and value of the variable. You can also provide tags (as a 
comma-separated list), but these are optional.

```console
$ var add --name username max_powers
```

If you do not provide a name or value, you will be prompted for them.

```console
$ var add

? Enter name: username
? Enter variable value: max_powers
```

## `get`

The `get` subcommand is a simple subcommand that retrieves and displays a stored variable. All of a variable's fields are 
displayed by the `get` subcommand.

```console
$ var get username
```

![Variable get output](../assets/var/get-output-var.svg)


## `update`

!!! tip "Aliases"
    `update` can also be called `edit`.

The update command allows you to make changes to a variable already stored in the database. Variables have few
editable fields, so you can only change the name or value of a variable.

```console
$ cb cmd update username --value guy_incognito

$ cb cmd update username --name new_username
```

!!! warning
    Be sure to wrap your value in quotes "" if it contains spaces.

Multiple fileds can be updated at the same time using the `--set` flag and key value pairs like `key=value`.

```console
$ cb cmd update username --set name=new_username --set value=guy_incognito
```
