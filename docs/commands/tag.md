# tag

The `tag` subcommand manages tags. Tags are labels you attach to commands and variables to
keep them organized. Once created, a tag can be attached to as many commands and variables
as you like, and you can filter your lists by tag to quickly find what you need.

The available subcommands for `tag` are:

- [add](#add)
- [get](#get)
- [update](#update)
- [list](#list)
- [search](#search)
- [delete](#delete)

Some of these subcommands also have aliases available. These will be discussed in the subcommands section.

---

## `add`

The `add` subcommand creates a new tag.

When creating a tag, you only have to provide a name. You will be prompted for the rest of the fields.

```console
> cb tag add docker
 
? Enter description: Commands related to Docker.
```

This same tag can be added in one line.

```console
> cb tag add docker "Commands related to Docker."
```

Notice that name and description are provided without a flag. They are not optional fields.

If you want to be prompted for every field, use the `--interactive` (or `-i`) flag.

```console
> cb tag add --interactive
 
? Enter name: docker
? Enter description: Commands related to Docker.
```

 
---

## `get`

The `get` subcommand retrieves a tag and displays all of its available fields.

```console
> cb tag get docker
```

![tag get output](../assets/tag/get-output.svg)
 
---

## `update`

!!! tip "Aliases"
`update` can also be called as `edit`.

The `update` subcommand makes changes to a tag you already have stored.

You can update a specific field by providing the appropriate flag and the new value.

```console
> cb tag update docker --description "Docker and container-related commands."
```

To rename a tag, use the `--name` flag. The tag will be updated across all commands and
variables that use it.

```console
> cb tag update docker --name containers
```

Multiple fields can be updated at once using the `--set` flag with `key=value` pairs.

```console
> cb tag update docker --set name=containers description="Docker and container-related commands."
```

!!! tip "Autocomplete"
Autocomplete is available for field names when using `--set`.

!!! warning
Be sure to wrap your values in quotes if they contain spaces.

If you want to update the current value of a field without supplying a completely new value,
use the `--edit` (or `-e`) flag. You will be prompted for each field, pre-filled with its
current value.

```console
> cb tag update docker --edit
 
? Enter name: docker
? Enter description: Docker and container-related commands.
```

If you only want to update a specific field in edit mode, use the `--edit-fields` (or `-ef`)
flag to specify which fields to prompt for.

```console
> cb tag update docker --edit --edit-fields description
 
? Enter description: Docker and container-related commands.
```

!!! warning
`--edit-fields` can only be used in conjunction with the `--edit` flag.
 
---

## `list`

!!! tip "Aliases"
`list` can also be called as `ls`.

The `list` subcommand displays all tags stored in your database.

```console
> cb tag list
```

![tag list output](../assets/tag/list-output.svg)

By default, only the name and description of each tag are displayed, and the default order
is by name. The default fields and ordering can be adjusted in your settings, or by supplying
additional options to the `list` subcommand.

To change the order, use the `--order` flag and specify the field you want to order by.

```console
> cb tag list --order description
```

To change the displayed fields, use the `--field` flag and specify the fields you want to display.

```console
> cb tag list --field name
```

If you have a large number of tags, use the `--limit` flag to cap the number of results.

```console
> cb tag list --limit 10
```

 
---

## `search`

!!! tip "Aliases"
`search` can also be called as `find`.

While `list` shows all of your tags, `search` lets you filter by the content of any available
field. By default, search looks across all fields.

```console
> cb tag search docker
```

![tag search output](../assets/tag/search-output-default.svg)

Using the `--in` flag, you can limit your search to specific fields.

```console
> cb tag search container --in description
```

![tag search output](../assets/tag/search-output-in-description.svg)

To control which fields appear in the results, use the `--field` flag.

```console
> cb tag search docker --field name
```

![tag search output](../assets/tag/search-output-field-name.svg)

As with `list`, you can limit the number of results using the `--limit` flag.

```console
> cb tag search docker --limit 5
```

 
---

## `delete`

!!! tip "Aliases"
`delete` can also be called as `rm`, `del`, or `remove`.

The `delete` subcommand removes a tag from your database. It only requires the name of the
tag you want to remove.

```console
> cb tag delete docker
```

!!! warning
Deleting a tag removes it from all commands and variables that use it. The commands and
variables themselves are not affected.
