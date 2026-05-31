# init

The `init` subcommand sets up shell integration for CmdBox. Once installed, shell integration
gives you access to the `cbe` command in addition to `cb`.

---

## `cb` vs `cbe`

By default, `cb` runs commands in a separate background process. This works for the vast
majority of use cases, but it means that certain commands cannot affect your current terminal
session. For example, activating a virtual environment with `cb activate` would work in the
background process and then immediately exit, leaving your current session unchanged.

The `cbe` command solves this. It runs a saved command directly in your current shell session,
so commands like activating a virtual environment, changing directories, or setting environment
variables work as expected.

```console
> cbe activate
```

`cbe` is only available after shell integration has been set up with `cb init`.

!!! note
    When using `cbe`, variable prompts are simplified due to stdout being captured by the shell. Features such as prompt
    history and tab completion are not available in this mode.

---

## Supported Shells

| Shell | Auto-install |
|---|---|
| PowerShell (`powershell`, `pwsh`) | Yes |
| Bash (`bash`) | Yes |
| Zsh (`zsh`) | Yes |
| Fish (`fish`) | Yes |
| CMD (`cmd`) | No, manual setup required |

---

## Setup

Run `cb init` to get started. If you do not specify a shell, CmdBox will attempt to detect
the one you are currently using.

```console
> cb init
```

By default this prints the integration snippet to your terminal along with instructions for
adding it to your shell profile manually. This is useful if you want to review what will be
added before making any changes.

To let CmdBox install the integration automatically, use the `--install` (or `-i`) flag:

```console
> cb init --install
```

You can also specify your shell explicitly:

```console
> cb init powershell --install
```

### Supported shell names

When specifying a shell, use one of the following names:

`bash`, `zsh`, `fish`, `powershell`, `pwsh`, `cmd`

---

## Custom Profile Path

By default CmdBox installs the integration into your shell's standard profile location. If
your profile is in a non-standard location, use the `--path` (or `-p`) flag to specify it:

```console
> cb init powershell --install --path C:\Users\Gorgoth\Documents\PowerShell\profile.ps1
```

---

## Backups

When `--install` modifies an existing profile file, CmdBox automatically creates a backup
before making any changes. The backup is saved alongside the original file with a `.bak`
extension. If the installation does not go as expected, you can restore your profile from
the backup manually.

---

## CMD Shell

CMD does not support the same profile-based integration as other shells. Running
`cb init cmd` will display a snippet and manual instructions for setting up the integration
yourself. Follow the printed instructions to complete the setup.

```console
> cb init cmd
```

---

## Verifying the Installation

After installing, open a new terminal session and run a saved command using `cbe`. If you
don't have any commands saved yet, create a quick test command first:

```console
> cb cmd add hello "echo CmdBox shell integration is working!"
```

Then run it with `cbe`:

```console
> cbe hello
CmdBox shell integration is working!
```

If the message is printed, your shell integration is set up correctly. You can delete the
test command afterward:

```console
> cb cmd delete hello
```

!!! warning
    You may need to restart your terminal or reload your profile manually.