package runtime

import (
	"fmt"
	"sort"
	"strings"
)

type RunContext struct {
	Cwd     string
	Env     map[string]string
	Capture bool
	Shell   string
	Timeout int // seconds - 0 means no timeout
	Emit    bool
	Verbose bool
}

func (ctx RunContext) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "cwd: %s\n", displayOrNone(ctx.Cwd))
	fmt.Fprintf(&b, "shell: %s\n", displayOrNone(ctx.Shell))

	if ctx.Timeout > 0 {
		fmt.Fprintf(&b, "timeout: %ds\n", ctx.Timeout)
	} else {
		fmt.Fprintf(&b, "timeout: none\n")
	}

	fmt.Fprintf(&b, "capture: %t\n", ctx.Capture)
	fmt.Fprintf(&b, "emit: %t\n", ctx.Emit)
	fmt.Fprintf(&b, "verbose: %t\n", ctx.Verbose)

	if len(ctx.Env) == 0 {
		fmt.Fprint(&b, "env: none")
		return b.String()
	}

	keys := make([]string, 0, len(ctx.Env))
	for k := range ctx.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprint(&b, "env:\n")
	for i, k := range keys {
		fmt.Fprintf(&b, "  %s=%s", k, ctx.Env[k])
		if i < len(keys)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func displayOrNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
