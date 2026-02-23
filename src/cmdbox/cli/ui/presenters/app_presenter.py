from cmdbox.cli.ui.primitives import banner


def render_version(version: str):
    return banner(
        f"CmdBox {version}",
        status="info",
    )
