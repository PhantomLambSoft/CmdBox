# CmdBox

A fast, structured, and searchable command runner for all terminal users.

CmdBox replaces fragile shell history and scattered notes with a clean, organized system for storing, searching, and 
executing commands.

It is designed for developers, sysadmins, power users, and anyone who wants a better terminal experience.

## Why CmdBox Exists
Many terminal users struggle with remembering and executing commands, relying on shell history, scattered notes, and 
repeated searches to recall commands. As workflow grows, this becomes inefficient and fragile.

CmdBox gives you:
- Named commands
- Parameterized templates
- Tagging
- Structured search
- Clean rich output
- Configurable field views
- Shell integration

## Quick Start
#### Basic commands

```bash
# Add a command called "deploy"
cb add deploy "git push origin main && fly deploy"

# Recall and run that command
cb run deploy

# List available commands
cb list

# Search for commands
cb search deploy
```

#### Use template variables
```bash
# Add a command with variable links
cb add ssh-server "ssh <username>@<host-one>"

# Add variables to store usernames, hosts, paths, etc.
cb var add username "maxpowers"
cb var add host-one "192.168.1.10"

# Command templates are populated before executing
cb run ssh-server
# What is executed:
ssh maxpowers@192.168.1.10
```

## Features
* Structured command storage
* Parameterized command templates and variables
* Field-based search and filtering
* Multi-line template execution
* Rich-based UI with configurable themes
* Configurable output views

## Installation
If you have pip installed:
```bash
pip install cmd-box
```

Or install from source:
```bash
git clone https://github.com/MalloyDelacroix/cmdbox.git
cd cmdbox
pip install -e .
```
