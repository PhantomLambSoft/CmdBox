# Installing CmdBox

Installing CmdBox can be done in one of two ways. Pick the section that fits you:

- **[I just want to install it (pip)](#option-1:-install-with-pip-recommended)**
- **[I want to install it from source](#option-2:-install-from-source)**

---

## Option 1: Install with pip (recommended)
pip is Python's built-in package manager. If you have Python installed, you have already have pip installed.
This is the fastest way to get CmdBox up and running. It handles everything for you, including making the `cb` command
available in your terminal.

### Step 1: Check if Python is installed
Open a terminal and run:

```console
python --version
```

CmdBox requires Python version **3.9 or higher**. If your version is higher than that, skip to step 2.

If you get an error, or your Python version is lower than 3.9, you'll need to install a compatible version of Python first.

<details>
<summary>
<strong>Installing Python on Windows</strong>
</summary>

1. Go to [python.org/downloads](python.org/downloads) and download the latest version for Windows.
2. Run the installer and follow the instructions.
3. **Important:** Make sure to check the "Add Python to PATH" option before clicking install. This step is easy to miss
and causes the most Python issues on Windows.
4. Once installed, close and reopen your terminal, then verify your installation:

```console
python --version
```
</details>

<details>
<summary>
<strong>Installing Python on MacOS</strong>
</summary>

1. Go to [python.org/downloads](python.org/downloads) and download the latest version for MacOS.
2. Run the `.pkg` file and follow the prompts.
3. Once installed, close and reopen your terminal, then run `python3 --version` to verify your installation.
> On MacOS, you may need to use `python3` and `pip3` instead of `python` and `pip` in all commands below.
</details>

<details>
<summary>
<strong>Installing Python on Linux</strong>
</summary>
Most Linux distributions come with Python pre-installed. To verify this open a terminal and run:

```bash
python3 --version
```
If your Linux distrubution does not have Python installed, or the Python version it does have is older than 3.9, run:

```bash
# Ubuntu / Debian
sudo apt update && sudo apt install python3 python3-pip

# Fedora
sudo dnf install python3 python3-pip
```
</details>

### Step 2: Install CmdBox

With Python installed, install CmdBox by running:

```console
pip install cmd-box
```

That's it! Once it finishes, CmdBox is available in any terminal window.

To verify your install, run the command `cb --version`.
You should see an output displaying the latest version of CmdBox.

**To update CmdBox later**

```console
pip install --upgrade cmd-box
```


## Option 2: Install from Source

Use this option if you want to contribute to CmdBox, run unreleased changes, or just prefer to work directly with the source code.

**Requirements:** Python 3.9 or higher (Python 3.14 recommended), [Git](https://git-scm.com/downlaods)

### 1. Clone the repository

```bash
git clone https://github.com/MalloyDelacroix/cmdbox.git
cd cmdbox
```

### 2. (Optional) Create a virtual environment

A virtual environment keeps CmdBox's dependencies isolated from the rest of your Python installation and is the 
recommended way to work on a Python project. However, if you install CmdBox only in the virtutal environment, that 
environment will have to be active in any terminal in which you wish to use CmdBox.

```bash
python -m venv .venv # or a name of your choice
```

Activate it:

```bash
# macOS and Linux
source .venv/bin/activate

# Windows
.venv\Scripts\activate
```

Your terminal prompt should now show `(venv)` at the start (or whatever name you chose for your virtual environment), indicating 
that the virtual environment is active.

### 3. Install in editable mode

```bash
pip install -e .
```

When using the `-e` flag, any changes you make to the source code will be reflected immediately without needing to reinstall.

### 4. Verify

```bash
cb --version
```

### Running tests
From the project root directory:

```bash
python -m unittest discover
```


## Verifying Your Installation

However you installed CmdBox, confirm it's working:

```console
cb --version
cb --help
```

---

## Uninstalling

| Install method | How to uninstall |
|----------------|------------------|
| pip | `pip uninstall cmd-box` |
| Source | Deactivate the virtual environment and delete the cloned folder |

> **Note:** uninstalling CmdBox will **not** remove any of the commands, variables, tags, or settings that it created. That data
is stored separately and will remain on your system. To remove this data, delete the CmdBox data folder from your system.

[comment]: # (TODO: provide link to dat folder doc)

---

## Need Help?

If something isn't working, [open an issue on github](https://github.com/MalloyDelacroix/cmdbox/issues). Include your
operating system, the output of `python --version`, and any error messages you receive.
