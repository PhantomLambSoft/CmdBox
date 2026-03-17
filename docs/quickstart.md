# Quickstart

This guide gets you from zero to running your first saved commands in under 5 minutes.

### 1. Save Your First Command
Pick a command you use regularly. For this example, we'll save a directory listing.

```console
cb cmd add list-home "ls -la ~"
```
You've just saved the command `ls -la ~` under the alias `list-home`.

### 2. Run It

```console
cb list-home
```

CmdBox looks up `list-home` and runs the saved command.

### 3. View Your Saved Commands

```console
cb cmd list
```

This lists every command you've saved, along with their alias and description.

### 4. Save a Command With a Variable
Variables make commands reusable. Use `<name>` as a placeholder:

```console
cb cmd add greet "echo Hello, <name>!"
```

```console
cb greet --name Max
```
Output: `Hello, Max!`

### 5. Remove a Command

```console
cb cmd remove list-home
```

That's the core loop. From here, explore:
- [The `cmd` reference](commands/cmd.md) for full command management options
- [Variables (`var`)](commands/var.md) to pre-set variable values for reuse
- [Tags (`tag`)](commands/tag.md) to organize your commaands and variables