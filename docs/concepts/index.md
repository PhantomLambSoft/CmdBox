# Concepts

Understanding these core ideas makes everything else in CmdBox click.

---

## Commands

A saved command is the heart of CmdBox. It holds the alias you use to recall, the template
that gets executed, and any additional information you want to store alongside it.

A command has several fields. Some are editable, and some are maintained automatically by CmdBox.

**Fields you can edit:**

| Field         | Description                                                  |
|---------------|--------------------------------------------------------------|
| `alias`       | The name you use to run the command.                         |
| `template`    | The shell instruction that gets executed.                    |
| `description` | A short description of what the command does.                |
| `tags`        | A list of tags used to categorize the command.               |
| `cwd`         | The working directory where the command is executed.         |
| `shell`       | The shell interpreter used to execute the command.           |
| `env`         | Environmental variables set for the duration of the command. |
| `timeout`     | The maximum number of seconds before the process is killed.  |

**Fields maintained by CmdBox:**

| Field          | Description                                              |
|----------------|----------------------------------------------------------|
| `date_created` | The date and time the command was first saved.           |
| `last_updated` | The date and time the command was last edited.           |
| `used`         | A count of how many times the command has been executed. |
| `last_used`    | The date and time the command was last executed.         |

These metadata fields are read-only. They can be displayed in `list` and `search` results or
used for sorting. See the [`cmd` reference](../commands/cmd.md) for more information.

### Alias

An alias is the short name you give a saved command. When you type `cb deploy`, CmdBox looks
up the alias `deploy` and runs whatever command is stored under it.

Aliases should be:

- **Short:** these are what you type to recall commands, so brevity matters.
- **Descriptive:** `deploy-prod` beats `dp`, especially when searching.
- **Lowercase with hyphens:** this is the conventional style, though aliases support any
  characters except whitespace.

### Template

A template is the actual shell instruction stored under an alias. It can be anything you would
normally type in a terminal: a single program, a pipeline, a long string of flags, or a
multi-line script.

```
alias:    git-graph
template: git log --oneline --graph --decorate --all
```

Multi-line templates are supported and are executed as a script. Behind the scenes, CmdBox
writes the template to a temporary file, executes it, then removes the file. See the
[`cmd` reference](../commands/cmd.md) for more information about multi-line templates.

---

## Variables

Variables make saved commands flexible and reusable. A variable is defined in a template by
placing the variable name inside angle brackets: `<variable-name>`. When the command runs,
CmdBox replaces each placeholder with a real value before executing.

Consider this template:

```
ssh <user>@<host> -p <port>
```

The same base template can connect to any server depending on the values supplied.

### Supplying variables at runtime

Pass variable values as flags when running the command:

```console
> cb ssh-connect --user admin --host 192.168.1.1 --port 22

ssh admin@192.168.1.1 -p 22
```

### Saving variable values

Variables can be saved to the database so they fill in automatically without being typed
each time:

```console
> cb var add user admin
> cb var add host 192.168.1.1
> cb var add port 22
```

Now the same command runs without any flags:

```console
> cb ssh-connect

ssh admin@192.168.1.1 -p 22
```

### Mixing saved and runtime values

A runtime flag always takes precedence over a saved value. This lets you override specific
variables on the fly while the rest fill in from saved values:

```console
> cb ssh-connect --host 10.0.0.5 --port 2222

ssh admin@10.0.0.5 -p 2222
```

### Variable resolution order

When a command runs, CmdBox resolves each variable in this order:

1. **Runtime flag** — a value passed directly with `--variable-name` always wins.
2. **Saved value** — a value stored via `cb var add` is used if no runtime flag is provided.
3. **Prompted** — if no value is found by either of the above, CmdBox will ask you to enter
   one before the command executes.

### Variable fields

Like commands, variables have metadata fields maintained by CmdBox:

| Field          | Description                                      |
|----------------|--------------------------------------------------|
| `date_created` | The date and time the variable was first saved.  |
| `last_updated` | The date and time the variable was last updated. |

See the [`var` reference](../commands/var.md) for full details on managing variables.

---

## Execution Context

Execution context controls the environment a command runs in. The four context fields are
`cwd`, `shell`, `env`, and `timeout`. All four are optional and default to `null`, meaning
CmdBox inherits them from your current terminal session when not set.

You can store context values directly on a command so they apply automatically every time
it runs:

```console
> cb cmd add deploy "npm run deploy" \
    --cwd "/home/user/projects/myapp" \
    --env NODE_ENV=production \
    --timeout 60
```

Now running `cb deploy` always uses that working directory, environment variable, and timeout
without any extra flags.

### Runtime overrides

The same flags are available on `cb run`. A value supplied at runtime always takes precedence
over the value stored on the command for that run, leaving the stored value unchanged for
future runs.

```console
> cb run deploy --env NODE_ENV=staging
```

This runs the command with `NODE_ENV=staging` for this invocation only. The stored value
remains `production`.

### Environment variable merging

Runtime `--env` flags merge with stored environment variables key-by-key rather than
replacing the stored dict entirely. Only the keys you supply at runtime are overridden.
Any stored keys you do not mention are preserved.

|             |                                    |
|-------------|------------------------------------|
| **Stored**  | `NODE_ENV=production`, `PORT=3000` |
| **Runtime** | `--env NODE_ENV=staging`           |
| **Result**  | `NODE_ENV=staging`, `PORT=3000`    |

### Timeout behavior

When a timeout is set, CmdBox manages the subprocess differently to ensure the entire process
tree is cleaned up when the limit is exceeded — not just the direct child process. The exit
code is set to `124` when a timeout occurs, consistent with the Unix `timeout` command
convention.

!!! note
    Execution context fields can be stored and updated via `cb cmd add` and `cb cmd update`.
    See the [`cmd` reference](../commands/cmd.md) for full details.

---

## Command References

A template can reference another saved command using the `<cmd:alias>` syntax. When the
command runs, CmdBox looks up the referenced alias and substitutes its template inline before
executing.

```
alias:    full-deploy
template: <cmd:git-pull> && <cmd:build> && <cmd:deploy>
```

Running `cb full-deploy` resolves each reference and executes the resulting command. This
lets you compose complex workflows from simpler saved commands without duplicating template
logic.

!!! tip
    Command references and variables can be used together in the same template.

---

## Tags

Tags are labels you attach to commands or variables for organization. A command or variable
can have multiple tags, and lists can be filtered by tag at any time.

```console
> cb cmd add deploy "git push origin main && fly deploy" --tags work,production
```

Filter your command list by tag:

```console
> cb cmd list --tag work
```

Tags themselves can be searched and managed via the [`tag` reference](../commands/tag.md).

---

## Settings

CmdBox behavior and appearance can be customized through settings. This includes defaults for
display fields, sort order, result limits, and more. Use the
[`settings` reference](../commands/settings.md) to view your current configuration and make
changes.

---

## Shell Integration

By default, `cb` runs commands in a background process separate from your current terminal
session. This works for most commands, but it means commands that need to affect your current
session, like activating a virtual environment or changing directories, will not work as
expected.

Shell integration adds the `cbe` command, which runs a saved command directly in your current
session. Setting it up is a one-time step handled by the [`init` subcommand](../commands/init.md).
