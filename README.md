# CmdBox

A fast, structured, and searchable command runner for the terminal.

CmdBox replaces fragile shell history and scattered notes with a clean, organized system for
storing, searching, and executing commands. Designed for anyone who works in the terminal,
from occasional users to seasoned developers.

---

## Why CmdBox?

Most terminal users have commands they run regularly. Some are short. Many are long and
complex, packed with flags and options that are easy to forget and tedious to type. Recalling them means digging
through shell history, hunting through notes, or searching online every time.

CmdBox gives every command a short, memorable alias. Run it instantly. No retyping, no
searching, no forgetting.

CmdBox isn't a replacement for shell scripting. It's a replacement for the growing pile of aliases and one-off functions
most terminal users accumulate for commands they run often but don't want to memorize, retype, or maintain as shell 
functions or scripts.

### How CmdBox Compares

|                                                                  | Alias | Shell function                | CmdBox                   |
|------------------------------------------------------------------|-------|-------------------------------|--------------------------|
| Named (non-positional) parameters                                | No    | No                            | Yes                      |
| Values shared across multiple commands                           | No    | No, unless duplicated         | Yes, via saved variables |
| Invocable from any shell without a matching per-shell definition | No    | No                            | Yes                      |
| Searchable, taggable                                             | No    | No                            | Yes                      |
| Scoped execution history with rerun                              | No    | No                            | Yes                      |
| Real scripting logic (loops, conditionals)                       | No    | Yes                           | No                       |
| No installation required                                         | Yes   | Yes                           | No                       |


See the [full comparison here](https://phantomlambsoft.github.io/CmdBox/comparisons/) for more details.

---

## Features

- Named commands with short, memorable aliases
- Parameterized templates with saved and runtime variables
- Stored execution context per command (working directory, shell, environment variables, and timeout) with runtime
  overrides
- Named profiles for separating commands, variables, and settings across contexts (work, personal, per-project)
- Command execution history with the ability to rerun past executions
- Tag-based organization and filtering
- Field-based search across commands, variables, and tags
- Multi-line template execution via script
- Rich terminal UI with configurable display fields

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

# Supply a different value at runtime to override a saved one
cb ssh-connect --host 192.168.1.1
```

#### Separate commands across profiles

```bash
# Create a profile
cb profile add work
# and switch to it
cb profile work

# Commands and variables created now belong to "work"
cb cmd add deploy "git push origin main && fly deploy"

# Switch back, "default" has its own separate set
cb profile default
cb cmd add deploy "npm run deploy"

# The same alias exists independently in each profile
cb profile work
cb deploy   # runs the work deploy
```

---

## Installation

```bash
pip install cmdbox-cli
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

---

## Documentation

Full documentation is available at [phantomlambsoft.github.io/CmdBox](https://phantomlambsoft.github.io/CmdBox).
