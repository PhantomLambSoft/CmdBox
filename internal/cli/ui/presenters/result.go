package presenters

import (
	"fmt"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/resolve"
	"github.com/PhantomLambSoft/CmdBox/internal/runtime"
)

func RenderExecutionResult(result runtime.ExecutionResult, title string, palette ui.Palette) string {
	ok := result.ExitCode == 0
	status := palette.Success
	if !ok {
		status = palette.Error
		title += " failed"
	}

	exitCodeText := fmt.Sprintf("Exit code: %d", result.ExitCode)
	banner := ui.Banner(title, exitCodeText, status, palette)
	codeBlock := ui.CodeBlock(result.Command, "Command", palette)

	parts := []string{banner, codeBlock}

	if strings.TrimSpace(result.Stdout) != "" {
		stdoutText := strings.Trim(result.Stdout, "\n")
		stdout := ui.SectionSuccess("STDOUT", stdoutText, palette)
		parts = append(parts, stdout)
	}
	if strings.TrimSpace(result.Stderr) != "" {
		stderrText := strings.Trim(result.Stderr, "\n")
		stderr := ui.SectionError("STDERR", stderrText, palette)
		parts = append(parts, stderr)
	}

	return ui.Stack(parts...)
}

func RenderPreviewResult(result resolve.Result, title string, showTrace bool, ctx runtime.RunContext, palette ui.Palette) string {
	resolvedCommand := ui.CodeBlock(result.Text, "Resolved Command", palette)
	renderedCtx := ui.CodeBlock(ctx.String(), "Run Context", palette)

	var stack string
	if showTrace && len(result.Trace) > 0 {
		trace := renderTraceSteps(result, palette)

		partWidth := ui.GetElementWidth(renderedCtx) + ui.GetElementWidth(trace)

		// If elements are wider than terminal, stack vertically
		if partWidth >= ui.GetTerminalWidth() {
			stack = ui.Stack(renderedCtx, trace)
		} else {
			stack = ui.StackHorz(renderedCtx, trace)
		}
	} else {
		stack = renderedCtx
	}

	body := ui.Stack(resolvedCommand, stack)

	return ui.Section(title, body, "", palette.Border, palette.PanelTitle)
}

var TraceStepColumns = []Column[resolve.TraceStep]{
	{Key: "kind", Header: "Kind", Render: func(s resolve.TraceStep, p ui.Palette) string {
		return p.TraceKind.Render(string(s.Kind))
	}},
	{Key: "key", Header: "Key", Render: func(s resolve.TraceStep, p ui.Palette) string {
		return p.TraceKey.Render(s.Key)
	}},
	{Key: "expanded_to", Header: "Expanded to", Render: func(s resolve.TraceStep, p ui.Palette) string {
		return p.TraceValue.Render(s.ExpandedTo)
	}},
}

func renderTraceSteps(result resolve.Result, palette ui.Palette) string {
	headers, rows := RenderTable(result.Trace, TraceStepColumns, palette)
	return ui.TablePanel("Trace", headers, rows, "", palette)
}
