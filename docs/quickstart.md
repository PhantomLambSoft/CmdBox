# Quickstart

This guide gets you from zero to running your first saved commands in under 5 minutes.

### 1. Save Your First Command
Pick a command you use regularly. For this example, we'll save a command that send 4 test pings to your own machine 
to verify that your network is working.

```console
cb cmd add ping-test "ping -n 4 127.0.0.1"
```
>Here, `ping-test` is the alias of your command, and `ping -n 4 127.0.0.1` is the command template. You will also be prompted 
to enter a description for your command and any tags that you want to associate with it. Both of these options can be left 
blank. Only an alias and command template are required.

You've just saved the command `ping -n 4 127.0.0.1` under the alias `ping-test`.

### 2. Run It

```console
cb ping-test
```

CmdBox looks up `ping-test` and runs the saved command.

### 3. View Your Saved Commands

```console
cb cmd list
```

This lists every command you've saved, along with their alias and description.

### 4. Save a Command With a Variable
Variables make commands reusable. They can be supplied at runtime, or saved along with commands to be reused in other commands.
For this example we'll supply the variable at runtime.

```console
cb cmd add greet "echo Hello, <name>!"
```

```console
cb greet --name Max
```
Output: `Hello, Max!`

### 5. Remove a Command

```console
cb cmd remove ping-test
```

That's the core loop. From here, explore:
- [The `cmd` reference](commands/cmd.md) for full command management options
- [Variables (`var`)](commands/var.md) to pre-set variable values for reuse
- [Tags (`tag`)](commands/tag.md) to organize your commaands and variables