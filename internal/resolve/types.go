package resolve

// CommandRecord represents a structure that holds a command alias and its associated template string.
type CommandRecord struct {
	Alias    string
	Template string
}

// VariableRecord represents a key-value pair where Name is the identifier and Value is the associated data.
type VariableRecord struct {
	Name  string
	Value string
}

// RefKind represents the kind of a reference, typically used to classify types like commands or variables.
type RefKind string

const (
	RefKindCommand  RefKind = "command"
	RefKindVariable RefKind = "variable"
)

// TraceStep represents a step in a trace sequence, detailing a reference kind, key, expansion, and an optional source.
type TraceStep struct {
	Kind       RefKind
	Key        string
	ExpandedTo string
	Source     string
}

// Result represents a resolution outcome with associated text and a sequence of trace steps detailing its derivation.
type Result struct {
	Text  string
	Trace []TraceStep
}
