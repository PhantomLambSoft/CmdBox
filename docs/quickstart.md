# Quick Start

This guide gets you from zero to running your first saved commands in under five minutes.

## 1. Save Your First Command

Pick a command you use regularly. For this example we'll save a command that sends four test
pings to your own machine to verify that your network is working.

```console
> cb cmd add ping-test "ping -n 4 127.0.0.1"
```

!!! note
    `ping-test` is the alias and `ping -n 4 127.0.0.1` is the command template. You will
    also be prompted to enter a description and any tags you want to associate with the
    command. Both can be left blank. Only an alias and template are required.

## 2. Run It

```console
> cb ping-test
```

CmdBox looks up `ping-test` and runs the saved command. Notice that no subcommand is needed,
just `cb` followed by the alias.

## 3. List Your Saved Commands

```console
> cb cmd list
```

This lists every command you have saved, along with its alias and description.

## 4. Save a Command With a Variable

Variables make commands reusable. They can be supplied at runtime or saved ahead of time so
they fill in automatically. For this example we'll supply the variable at runtime.

```console
> cb cmd add greet "echo Hello, <name>!"
```

```console
> cb greet --name Beautiful
```

```console
Hello, Beautiful!
```

## 5. Delete a Command

```console
> cb cmd delete ping-test
```

---

That's the core loop. From here, explore:

- [The `cmd` reference](commands/cmd.md) for full command management options
- [Variables (`var`)](commands/var.md) to save variable values for reuse across commands
- [Tags (`tag`)](commands/tag.md) to organize your commands and variables
- [Concepts](concepts.md) for a deeper understanding of how CmdBox works
