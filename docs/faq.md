# FAQ

---

## Installation and Setup

### What Python version do I need?

CmdBox requires Python 3.10 or higher. To check your current version:

```console
> python --version
```

### How do I know if the installation worked?

Run the following command after installing:

```console
> cb --version
```

If CmdBox is installed correctly, the current version number will be printed. If you see an
error, make sure that your Python scripts directory is included in your system's PATH.

### Where does CmdBox store my data?

This location differs depending on your operating system:
- On Windows, CmdBox stores data in `C:/Users/<username>/AppData/Roaming/PhantomLamb/cmdbox/`.
- On macOS, CmdBox stores data in `~/Library/Application Support/PhantomLamb/cmdbox/`.
- On Linux, CmdBox stores data in `~/.local/share/PhantomLamb/cmdbox/`.

---

## Basic Usage

### What is the difference between `cb git-graph` and `cb run git-graph`?

They do the same thing. Typing `cb <alias>` directly is a shorthand that CmdBox provides for
convenience. The `run` subcommand is available when you need additional options, such as
setting a working directory or previewing the resolved command before executing it. See the
[`run` reference](commands/run.md) for details.

### What happens if my alias matches a system command?

CmdBox aliases only activate when invoked through `cb`. Typing the alias name alone in your
terminal will still run whatever your system resolves it to. Your system is not affected.

### Can I use CmdBox inside a shell script?

Yes. `cb <alias>` is a standard shell command and works anywhere a command can be called,
including inside batch files and scripts.

---

## Variables

### What happens if I don't supply a variable when running a command?

If a variable is not supplied at runtime and no saved value exists for it, CmdBox will prompt
you to enter a value before the command executes. See [Concepts](../concepts.md) for a full
explanation of variable resolution.

### Can I use the same variable in multiple commands?

Yes! A saved variable value is global and will be used by any command template that contains
a placeholder with the same name. For example, saving `cb var add host 10.0.0.5` means every
command that contains `<host>` will use that value automatically.

---

## Data and Safety

### How do I back up my saved commands?

You can copy your data folder to a safe location. Your commands, variables, tags, and history are
stored in an SQLite database file that can be copied like any other file. Your settings are
stored separately in a `.toml` file, which is plain text and equally straightforward to copy.

### If I uninstall CmdBox, will I lose my saved commands?

No. Uninstalling CmdBox removes the application but leaves your data folder intact. If you
reinstall CmdBox later, your saved commands, variables, tags, and history will still be there.

### Can I share my commands with someone else?

Yes. Your data is stored in an SQLite database file. This file can be sent to another CmdBox
user and placed in their data folder, giving them access to all of your saved commands,
variables, and tags.

---

## Troubleshooting

### CmdBox is not recognized after installing

This usually means your Python scripts directory is not in your system PATH. To fix this,
find the location of your Python scripts folder:

```console
> python -c "import sysconfig; print(sysconfig.get_path('scripts'))"
```

Then add that path to your system's PATH environment variable. On Windows, you can do this
through **System Properties > Environment Variables**.

### A command runs but variables are not being replaced

Make sure the variable name in your template exactly matches the name of the saved variable
or the runtime flag you are supplying. Variable names are case-sensitive. For example,
`<Host>` and `<host>` are treated as different variables.

### My settings file is not loading correctly

If CmdBox fails to start after editing the settings file externally, the file may be
malformed. You can reset your settings to their defaults by deleting the settings file from
your data folder. CmdBox will generate a fresh one with default values on the next run.