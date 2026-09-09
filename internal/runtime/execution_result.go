package runtime

// ExecutionResult represents the outcome of a command execution, including command details, status, and output logs.
type ExecutionResult struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}
