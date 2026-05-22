# settings

The `settings` subcommand lets you view and edit your CmdBox configuration. Settings control
default behaviors across the app (things like default display fields, ordering, and limits)
so you don't have to supply the same options every time you run a command.

The available subcommands for `settings` are:

- [show](#show)
- [edit](#edit)

---

## `show`

The `show` subcommand displays your current settings in a table.

```console
> cb settings show
```

![settings show output](../assets/settings/show-output.svg)

To display only specific settings, use the `--fields` flag with a comma-separated list of
field names.

```console
> cb settings show --fields editor,data_file
```

---

## `edit`

The `edit` subcommand opens your settings for editing. By default, settings are edited
interactively in the terminal. The current settings will be displayed with their current values pre-filled.
You can edit them directly in the terminal.

```console
> cb settings edit
```

!!! tip
When editing settings interactively in the terminal, press `Esc` to exit edit mode without saving changes, or 
press `Ctrl+S` to save the changes and exit edit mode.

If you prefer to edit the settings file directly, use the `--external` (or `-e`) flag.
This opens the settings file in your default text editor.

```console
> cb settings edit --external
```

!!! warning
When using `--external`, take care not to change field names or the file structure.
CmdBox may not start correctly if the settings file is malformed.

!!! tip
Not sure which editor will open? You can set your preferred editor from within the
interactive settings editor itself.