## CmdBox vs. Aliases and Shell Scripts

Shell aliases and scripts can already do a lot, and if all you need is `alias=command`, they're a perfectly good,
dependency free solution. CmdBox exists for the point past that: commands that take parameters, need to work the same
way across shells, or that you want to find again later without remembering exactly what you named them.

Here's a direct comparison.

### Named variables, not positional arguments

An alias can't take arguments at all. A shell function can, but only positionally, so every value has to be supplied in
the exact order the function expects, with no labels. Consider a long command with lots of arguments:

```bash
pg-dev() {
  docker run -d --name "$1" -e POSTGRES_USER="$2" -e POSTGRES_PASSWORD="$3" \
    -e POSTGRES_DB="$4" -p "${5}:5432" -v "${6}:/var/lib/postgresql/data" \
    --network "$7" "postgres:${8}"
}
```

```console
> pg-dev pg-dev admin s3cret myapp 5433 ~/pgdata devnet 16
```

Eight bare values in a row. Mix up any two and nothing catches it, the container just starts with the wrong password
in the wrong network.

```console
> cb cmd add pg-dev "docker run -d --name <container-name> -e POSTGRES_USER=<db-user> \
  -e POSTGRES_PASSWORD=<db-password> -e POSTGRES_DB=<db-name> -p <host-port>:5432 \
  -v <volume-path>:/var/lib/postgresql/data --network <network-name> postgres:<pg-version>"

> cb pg-dev --db-user admin --db-password s3cret --db-name myapp --host-port 5433 \
  --volume-path ~/pgdata --network-name devnet --pg-version 16
```

Every value is labeled, in any order, and self-documenting when you read it back. If you forget to enter one, it doesn't 
just fail, it prompts you for what you missed.

### Saved values shared across commands

If you supply a shell function with default values, they live inside that one function. If three different functions all
need the same database host, you're hardcoding it three times, and updating it means editing three files.

With CmdBox:

```console
> cb var add host 10.0.0.5
```

Any command template containing `<host>` picks up that value automatically. Update it once with
`cb var update host --value new-value` and every command that references it now uses the new value on its next run.

### Reachable from any shell, not tied to one

A bash alias or function only exists inside a bash session. It's defined in `.bashrc`, and it simply isn't there if you'
re sitting in PowerShell or cmd. PowerShell and cmd each have their own, incompatible mechanisms for the same idea (
`Set-Alias`, batch files), so replicating one saved command across bash, zsh, fish, PowerShell, and cmd means writing
and maintaining a separate definition in each.

CmdBox itself is a single cross-platform program, so a command saved once is reachable by name from any of those shells,
without needing a matching alias defined in each one's own config file.

!!! note
    CmdBox does not translate a command's syntax between shells. A template using bash-specific
    constructs still needs bash installed, and either needs to be called from bash or have
    `--shell bash` specified, either at runtime or saved with the command.

    What CmdBox does do is decouple the shell you're typing in from the shell the command executes
    with. A plain external-program invocation, like the Postgres example above, has no shell-specific
    syntax at all, so it runs identically no matter which shell resolves it, as long as the program
    itself is installed.

### No file per command

A script needs a file: created, made executable, placed somewhere on `PATH`. That's manageable for one or two, but it
adds up, and organizing dozens of scripts across a `bin` directory has no built-in search or listing beyond `ls`.

```console
> cb cmd add deploy "npm run deploy" --cwd "/home/user/projects/myapp"
```

One line, no file, no `chmod`, no `PATH` entry. It's immediately searchable and taggable:

```console
> cb cmd search deploy
> cb cmd list --tag prod
```

### Scoped, structured history

Bash's `history | grep` searches everything you've ever typed in that shell, unstructured. CmdBox's history only logs
runs made through CmdBox, tied to the alias, the variables used, and the exit code, so `cb history show 3` tells you
exactly what ran and with what values, and `cb history rerun 3` repeats it exactly.

### Where scripts and functions still win

To be fair, this isn't one-sided. If you need real programming logic (loops, conditionals, calling out to other programs
and parsing their output) that's a shell script's job, not CmdBox's. CmdBox stores and resolves command templates, it
isn't a scripting language. For a one-off alias you'll only ever use in one shell and never share, a plain `alias` is
still the lightest possible tool for that specific job.

### Summary

|                                                                  | Alias | Shell function                | CmdBox                   |
|------------------------------------------------------------------|-------|-------------------------------|--------------------------|
| Named (non-positional) parameters                                | No    | No                            | Yes                      |
| Values shared across multiple commands                           | No    | No, unless duplicated         | Yes, via saved variables |
| Invocable from any shell without a matching per-shell definition | No    | No                            | Yes                      |
| Searchable, taggable                                             | No    | No                            | Yes                      |
| Scoped execution history with rerun                              | No    | No                            | Yes                      |
| Real scripting logic (loops, conditionals)                       | No    | Yes                           | No                       |
| No installation required                                         | Yes   | Yes                           | No                       |

CmdBox isn't a replacement for shell scripting. It's a replacement for the growing pile of aliases and one-off functions
most terminal users accumulate for commands they run often but don't want to memorize, retype, or maintain as individual
functions or scripts.