# Quoting and Templates

When you save a command with CmdBox, the exact text stored as the template depends on *how* you entered it: as an inline
argument or through the interactive prompt. Understanding this difference will save you a debugging session down the
road.

## The short version

- **Inline argument** (`cb cmd add my-cmd "echo hello world"`): your shell parses the line first. Quotes used to group words
  into a single argument are stripped before CmdBox ever sees them. CmdBox stores the plain text.

- **Interactive prompt**: there is no shell involved. Whatever characters you type are stored exactly as typed,
  including any quote characters.


Neither behavior is a bug. They're just two different input paths with two different rules, and it's worth knowing which
one you're using.

## Why this happens

When you run a command in a terminal, the shell (PowerShell, cmd, bash, zsh) reads the full line and breaks it into
arguments *before* handing them to the program you're running. Quotes are one of the tools the shell uses to do that
grouping. For example:

```
cb cmd add echo-test "echo hello world"
```

Here, `"echo hello world"` tells the shell "treat these three words as one argument." The shell removes the quote
characters as part of that parsing. By the time CmdBox receives the `template` argument, it is just:

```
echo hello world
```

No quotes, because the shell already did its job.

Compare that to the interactive prompt:

```
cmdbox> cb cmd add echo-test
Template: "echo hello world"
```

There's no shell parsing happening at this point. Prompt input is read as raw text, so if you type quote characters,
they are saved as literal characters in the template:

```
"echo hello world"
```

## Why it matters

CmdBox runs single-line templates directly, and multi-line templates by writing them to a temporary script file. Either
way, the template is what gets executed, quotes and all. This means the two entry paths above will actually run
differently:

| Entry method                              | What gets stored     | What happens when run                                                                                  |
|-------------------------------------------|----------------------|--------------------------------------------------------------------------------------------------------|
| `cb cmd add echo-test "echo hello world"` | `echo hello world`   | Runs as three separate words passed to `echo`                                                          |
| Prompted, typed `"echo hello world"`      | `"echo hello world"` | The quotes are passed along to the shell at execution time, which can change how the words are grouped |

If you want a command to behave a specific way when it runs, especially one involving multiple words that should be
treated as a single value, think about how it would behave if you typed it directly into your terminal. CmdBox does not
add or remove quoting on your behalf. What you provide (after your shell, if any, has processed it) is what gets stored
and later executed.

## Getting the result you want

If your goal is for a template to run and print `hello world` as it would if you typed it directly into a terminal,
enter it inline exactly as you would type it yourself:

```
cb cmd add echo-test "echo hello world"
```

The outer quotes here exist only so your shell treats the whole phrase as a single argument to `cb cmd add`, since
`alias` and `template` are separate positional arguments. Once stored, the template is the plain text
`echo hello world`, and it will run the same way it would if you typed that line yourself.

If you need the *stored template itself* to contain quote characters, for example to group words together at execution
time, add them as part of the phrase rather than around the whole thing:

```
cb cmd add echo-test "echo 'hello world'"
```

Here, the outer double quotes are stripped by your shell as usual, and the inner single quotes survive because they're
part of the argument text, not shell syntax. The stored template becomes:

```
echo 'hello world'
```

which runs and prints `hello world` on a single line.

## Rule of thumb

Whatever text your shell would hand to a program on the command line is the same text CmdBox will store and later
execute. If you're not sure how something will behave, try typing it directly into your terminal first. If it does what
you expect there, entering it the same way as an inline template argument will produce the same stored (and executed)
result.