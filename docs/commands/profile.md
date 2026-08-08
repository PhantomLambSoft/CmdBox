# profile

The `profile` subcommand manages named profiles. Profiles let you switch between separate sets of commands, variables,
and settings, useful for separating work and personal commands, keeping the same aliases across different projects, or
maintaining different variable values across environments.

The available subcommands for `profile` are:

- [add](#add)
- [get](#get)
- [update](#update)
- [list](#list)
- [delete](#delete)
- [status](#status)
- [switch](#switch)
- [cmd](#cmd)
- [var](#var)
- [settings](#settings)

---

## `add`

Creates a new profile.

```console
> cb profile add work

? Enter description: Work-related commands and variables.
```

```console
> cb profile add work "Work-related commands and variables."
```

If you want to be prompted for every field, use the `--interactive` (or `-i`) flag.

---

## `get`

Retrieves and displays a saved profile.

```console
> cb profile get work
```

---

## `update`

!!! tip "Aliases"
    `update` can also be called as `edit`.

Updates an existing profile.

```console
> cb profile update work --description "New description."
```

To rename a profile, use the `--name` flag.

```console
> cb profile update work --name office
```

!!! warning 
    The `default` profile cannot be renamed. It is the one profile guaranteed to always exist, and its name
    determines which settings file is used. Attempting to rename it raises an error.

Multiple fields can be updated at once using `--set` with `key=value` pairs, or use `--edit` (or `-e`) to be prompted
interactively for each field, pre-filled with its current value. These work the same way as they do for [
`cmd update`](cmd.md#update)
and [`var update`](var.md#update).

---

## `list`

!!! tip "Aliases"
    `list` can also be called as `ls`.

Displays all stored profiles.

```console
> cb profile list
```

---

## `delete`

!!! tip "Aliases"
    `delete` can also be called as `rm`, `del`, or `remove`.

Deletes a profile.

```console
> cb profile delete work
```

If the profile still has commands or variables assigned to it, the delete is blocked and the error tells you how many.
Use `--force` to delete the profile along with everything in it.

```console
> cb profile delete work --force
```

!!! warning 
    The `default` profile can never be deleted, regardless of `--force`. A profile that is currently active (for
    commands, variables, or settings) also cannot be deleted, switch away from it first.

---

## `status`

!!! tip "Aliases"
    `status` can also be called as `where`.

Shows the currently active profile for commands, variables, and settings. This is the "where am I" check, especially
useful once you start switching profiles independently, since it's easy to lose track of which profile applies to what.

```console
> cb profile status
```

The output flags whether all three are currently linked (pointing at the same profile) or have diverged.

---

## `switch`

!!! tip "Aliases"
    `switch` can also be called as `set`.

Switches the active profile. This is what runs when you use the shorthand
`cb profile <name>`, you rarely need to type `switch` explicitly.

```console
> cb profile work
```

With no flags, this switches command, variable, and settings profiles all at once, the default "linked" behavior. To
switch only some of them, use
`--cmd`, `--var`, and/or `--settings`:

```console
> cb profile work --cmd --var
```

This switches the command and variable profiles to `work`, leaving whatever settings profile was active untouched.

!!! note 
    Because the bare-name shorthand only works when the name doesn't collide with a real `profile` subcommand, a profile 
    literally named `add`, `list`, `status`, etc. can't be reached via `cb profile <name>`. Use `cb profile switch <name>` 
    explicitly in that case.

---

## `cmd`

Switches only the active command profile.

```console
> cb profile cmd work
```

---

## `var`

Switches only the active variable profile.

```console
> cb profile var personal
```

---

## `settings`

Switches only the active settings profile.

```console
> cb profile settings work
```