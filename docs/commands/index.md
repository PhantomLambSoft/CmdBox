# Command Reference

CmdBox is organized into six subcommands. Each handles a distinct area of functionality.

| Subcommand                | Description |
|---------------------------|---|
| [`cmd`](cmd.md)           | Add, edit, search, and delete saved commands. |
| [`var`](var.md)           | Add, edit, search, and delete saved variables. |
| [`tag`](tag.md)           | Add, edit, search, and delete tags. |
| [`run`](run.md)           | Execute saved commands by alias. |
| [`init`](init.md)         | Set up shell integration to enable running commands in your current terminal session. |
| [`history`](history.md)   | View and re-run past command executions. |
| [`settings`](settings.md) | View and edit your CmdBox configuration. |

---

## Running a Saved Command

The most common thing you will do in CmdBox is run a saved command. You do not need the `run`
subcommand for this — just use the alias directly:

```console
> cb git-graph
```

The `run` subcommand is available when you need more control over how a command executes, such
as setting a working directory, passing environment variables, or previewing the resolved
command before running it:

```console
> cb run git-graph --cwd C:\Projects\myapp --verbose
```

See the [`run` reference](run.md) for the full list of options.

---

## Getting Help

Every subcommand supports the `--help` flag, which displays its available options and usage
directly in the terminal:

```console
> cb --help
> cb cmd --help
> cb cmd add --help
```