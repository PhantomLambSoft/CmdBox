package runtime

type RunContext struct {
	Cwd     string
	Env     map[string]string
	Capture bool
	Shell   string
	Timeout int // seconds - 0 means no timeout
	Emit    bool
	Verbose bool
}
