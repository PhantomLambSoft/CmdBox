# CmdBox

**Source Code:** [github.com/PhantomLambSoft/CmdBox](https://github.com/PhantomLambSoft/CmdBox)

CmdBox is a fast, structured, and searchable command runner for the terminal.

CmdBox replaces fragile shell history, scattered notes, and online searches with an organized
system for storing, searching, and executing commands. It is designed for anyone who works in the
terminal, from occasional users to seasoned developers.

---

## Why CmdBox?

Most terminal users have commands they run regularly. Some are short. Many are long and complex,
packed with flags and options that are easy to forget. Recalling them means digging through
shell history, hunting through notes, or searching online every time.

CmdBox gives every command a short, memorable alias. Run it instantly. No retyping, no
searching, no forgetting.

---

## Features

- **Named commands:** save any command under a short alias and run it anytime
- **Parameterized templates:** use variables to make a single command work across many situations
- **Command history:** every execution is recorded so you can review and rerun past commands
- **Tagging:** organize commands and variables with tags and filter by them at any time
- **Field-based search:** find commands, variables, and tags by searching any field
- **Multi-line script support:** save and run multi-line commands as scripts
- **Rich terminal UI:** clean, colorized output with configurable display fields

---

## Quick Start

#### Save and run a command

```bash
# Save a command under an alias
cb cmd add git-graph "git log --oneline --graph --decorate --all"

# Run it by alias, no subcommand needed
cb git-graph

# List all saved commands
cb cmd list

# Search saved commands
cb cmd search git
```

#### Use variables for flexible commands

```bash
# Save a command with variable placeholders
cb cmd add ssh-connect "ssh <user>@<host> -p <port>"

# Save variable values so they fill in automatically
cb var add user admin
cb var add host 10.0.0.5
cb var add port 22

# Run the command, variables are resolved before executing
cb ssh-connect

# What gets executed:
ssh admin@10.0.0.5 -p 22

# Supply a different value at runtime to override the saved one
cb ssh-connect --host 192.168.1.1
```

---

## Installation

```bash
pip install cmdbox
```

Or install from source:

```bash
git clone https://github.com/PhantomLambSoft/CmdBox.git
cd cmdbox
pip install .
```

After installation, verify it worked:

```bash
cb --version
```