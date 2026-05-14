# var

The `var` subcommand manages variables. Variables are reusable values that can be referenced
inside your saved commands using the `<variable-name>` syntax, allowing a single command to
work across many different situations.

The available subcommands for `var` are:

- [add](#add)
- [get](#get)
- [update](#update)
- [list](#list)
- [search](#search)
- [delete](#delete)
- [tag](#tag)
- [untag](#untag)


  Some of these subcommands also have aliases available. These will be discussed in the subcommands section.

---

## `add`

The `add` subcommand saves a new variable to your CmdBox database.

When creating a variable, you only have to provide a name. You will be prompted for the rest of the fields.

```console
> cb var add host
 
? Enter value: 192.168.1.1
? Enter tags (comma-separated): network,work
```

This same variable can be added in one line.

```console
> cb var add host 192.168.1.1
```

Notice that name and value are provided without a flag. They are not optional fields.

You will always be prompted for tags.

!!! tip "Autocomplete"
Autocomplete is available for stored tags. In the tag prompt, start typing and the available tags will be suggested.

If you want to be prompted for every field, use the `--interactive` (or `-i`) flag.

```console
> cb var add --interactive
 
? Enter name: host
? Enter value: 192.168.1.1
? Enter tags (comma-separated): network,work
```

 
---

## `get`

The `get` subcommand retrieves a variable and displays all of its available fields along with its tags.

```console
> cb var get host
```

![var get output](../assets/var/get-output.svg)
 
---

## `update`

!!! tip "Aliases"
`update` can also be called as `edit`.

The `update` subcommand is used to make changes to a variable you already have stored.

You can update a specific field by providing the appropriate flag and the new value.

```console
> cb var update host --value 10.0.0.5
```

To rename a variable, use the `--name` flag.

```console
> cb var update host --name server-host
```

Multiple fields can be updated at once using the `--set` flag with `key=value` pairs.

```console
> cb var update host --set value=10.0.0.5 name=server-host
```

!!! tip "Autocomplete"
Autocomplete is available for field names when using `--set`.

!!! warning
Be sure to wrap your values in quotes if they contain spaces.

If you want to update the current value of a field without supplying a completely new value, use the `--edit`
(or `-e`) flag. You will be prompted for each field, pre-filled with its current value.

```console
> cb var update host --edit
 
? Enter name: host
? Enter value: 10.0.0.5
```

If you only want to update a specific field in edit mode, use the `--edit-fields` (or `-ef`) flag to specify
which fields to prompt for.

```console
> cb var update host --edit --edit-fields value
 
? Enter value: 10.0.0.5
```

!!! warning
`--edit-fields` can only be used in conjunction with the `--edit` flag.
 
---

## `list`

!!! tip "Aliases"
`list` can also be called as `ls`.

The `list` subcommand displays all variables stored in your database.

```console
> cb var list
```

![var list output](../assets/var/list-output.svg)

By default, only the name, value, and tags of each variable are displayed, and the default order is by name.
The default fields and ordering can be adjusted in your settings, or by supplying additional options to the `list`
subcommand.

To change the order, use the `--order` flag and specify the field you want to order by.

```console
> cb var list --order value
```

To change the displayed fields, use the `--field` flag and specify the fields you want to display.

```console
> cb var list --field name --field value
```

If you have a large number of variables, use the `--limit` flag to cap the number of results.

```console
> cb var list --limit 10
```

List can also be filtered to only variables that have a specific tag.

```console
> cb var list --tag network
```

The `--tag` flag can be used multiple times to filter by multiple tags.

```console
> cb var list --tag network --tag work
```

!!! tip
When using multiple `--tag` flags, variables that feature any of those tags will be displayed.
 
---

## `search`

!!! tip "Aliases"
`search` can also be called as `find`.

While `list` lets you filter variables by tag, `search` lets you filter by the content of any available field.
By default, search looks in the name and value fields.

```console
> cb var search 192
```

![var search output](../assets/var/search-output-default.svg)

Using the `--in` flag, you can limit your search to specific fields.

```console
> cb var search work --in tags
```

![var search output](../assets/var/search-output-in-tags.svg)

To control which fields appear in the results, use the `--field` flag.

```console
> cb var search 192 --field name
```

![var search output](../assets/var/search-output-field-name.svg)

As with `list`, you can limit the number of results using the `--limit` flag.

```console
> cb var search 192 --limit 3
```

 
---

## `delete`

!!! tip "Aliases"
`delete` can also be called as `rm`, `del`, or `remove`.

The `delete` subcommand removes a variable from your database. It only requires the name of the variable
you want to remove.

```console
> cb var delete host
```

!!! warning
This action cannot be undone.
 
---

## `tag`

The `tag` subcommand adds tags to a stored variable. Provide the name of the variable followed by the
tag you want to add.

```console
> cb var tag host network
```

!!! tip "Autocomplete"
Autocomplete is available for both variable names and tag names.
 
---

## `untag`

The `untag` subcommand removes a tag from a stored variable. It works the same way as `tag`.

```console
> cb var untag host network
```

