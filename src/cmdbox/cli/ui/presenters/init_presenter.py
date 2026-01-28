from cmdbox.cli.ui.presenters.init_instructions import get_instructions
from cmdbox.cli.ui.primitives import (
    section,
    to_text,
    stack,
    spacer,
    bullet_list,
    banner,
)


def render_install_instructions(
    install_snippet: str,
    shell: str,
    *,
    title: str = "Installation Instructions",
    include_help_text: bool = True,
):
    snippet = to_text(install_snippet, style="code.block")
    instruction_text = get_instructions(shell)
    instructions = section(
        title="What to do with this snippet",
        body=bullet_list(instruction_text, item_style="ui.muted"),
        border_style="status.info",
    )
    if include_help_text:
        help_text = section(
            title="Automate",
            body=to_text(
                "Re-run the `init` command with the `--install` flag to automatically install the snippet into your configuration file",
                style="ui.muted",
            ),
        )
    else:
        help_text = ""
    body = stack(snippet, spacer(0), instructions, spacer(0), help_text)
    return section(
        title=title,
        body=body,
        border_style="ui.border",
    )


def render_install_success():
    return banner(
        title="Shell integration installed",
        subtitle="Restart your shell for changes to take effect",
        status="success",
    )


def render_shell_output(shell: str):
    return section(
        title="Shell detected",
        body=to_text(shell, style="code.inline"),
        border_style="status.info",
    )
