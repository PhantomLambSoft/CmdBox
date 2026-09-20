package handlers

type RawRunContext struct {
	Cwd        string
	HasCwd     bool
	Env        []string
	Capture    bool
	Shell      string
	HasShell   bool
	Timeout    int
	HasTimeout bool
	Emit       bool
	Verbose    *bool // tri-state
}

func RunRunCommand(alias string, runtimeVars map[string]string, ctx RawRunContext, profile string) error {
	return nil
}

func RunPreviewCommand(alias string, runtimeVars map[string]string, ctx RawRunContext, profile string) error {
	return nil
}
