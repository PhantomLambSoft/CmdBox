package resolve

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

// fakeLookup is an in-memory Lookup implementation for resolver tests.
type fakeLookup struct {
	commands  map[string]CommandRecord
	variables map[string]VariableRecord

	// cmdErr/varErr, if set for a given key, are returned verbatim instead of the
	// normal not-found error - used to simulate unexpected lookup failures.
	cmdErr map[string]error
	varErr map[string]error
}

func newFakeLookup() *fakeLookup {
	return &fakeLookup{
		commands:  make(map[string]CommandRecord),
		variables: make(map[string]VariableRecord),
		cmdErr:    make(map[string]error),
		varErr:    make(map[string]error),
	}
}

func (f *fakeLookup) GetCommand(alias string) (CommandRecord, error) {
	if err, ok := f.cmdErr[alias]; ok {
		return CommandRecord{}, err
	}
	rec, ok := f.commands[alias]
	if !ok {
		return CommandRecord{}, fmt.Errorf("looking up command alias: %w", repository.ErrUnknownAlias)
	}
	return rec, nil
}

func (f *fakeLookup) GetVariable(name string) (VariableRecord, error) {
	if err, ok := f.varErr[name]; ok {
		return VariableRecord{}, err
	}
	rec, ok := f.variables[name]
	if !ok {
		return VariableRecord{}, fmt.Errorf("looking up variable name: %w", repository.ErrUnKnownName)
	}
	return rec, nil
}

func rootLabel(s string) *string { return &s }

// --- Resolve: plain text / basic substitution ---

func TestResolvePlainTextNoRefs(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("just plain text", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "just plain text" {
		t.Fatalf("Text = %q, want %q", res.Text, "just plain text")
	}
	if len(res.Trace) != 0 {
		t.Fatalf("Trace = %v, want empty", res.Trace)
	}
}

func TestResolveNilRootLabelDefaultsToInput(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["name"] = VariableRecord{Name: "name", Value: "world"}
	r := NewResolver(lookup, false, 10)

	got, err := r.Resolve("hello <name>", nil, nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	want, err := r.Resolve("hello <name>", rootLabel("<input>"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Resolve() with nil rootLabel = %+v, want same result as explicit %q label: %+v", got, "<input>", want)
	}
}

func TestResolveVariableFromStore(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["name"] = VariableRecord{Name: "name", Value: "world"}
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("hello <name>", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "hello world" {
		t.Fatalf("Text = %q, want %q", res.Text, "hello world")
	}
	if len(res.Trace) != 1 {
		t.Fatalf("Trace = %v, want 1 step", res.Trace)
	}
	want := TraceStep{Kind: RefKindVariable, Key: "name", ExpandedTo: "world", Source: "stored"}
	if res.Trace[0] != want {
		t.Fatalf("Trace[0] = %+v, want %+v", res.Trace[0], want)
	}
}

func TestResolveVariableFromRuntimeVarsTakesPrecedence(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["name"] = VariableRecord{Name: "name", Value: "stored-value"}
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("hello <name>", rootLabel("root"), map[string]string{"name": "runtime-value"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "hello runtime-value" {
		t.Fatalf("Text = %q, want %q", res.Text, "hello runtime-value")
	}
	want := TraceStep{Kind: RefKindVariable, Key: "name", ExpandedTo: "runtime-value", Source: "runtime"}
	if len(res.Trace) != 1 || res.Trace[0] != want {
		t.Fatalf("Trace = %+v, want [%+v]", res.Trace, want)
	}
}

func TestResolveCommand(t *testing.T) {
	lookup := newFakeLookup()
	lookup.commands["build"] = CommandRecord{Alias: "build", Template: "go build"}
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("run: <cmd:build>", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "run: go build" {
		t.Fatalf("Text = %q, want %q", res.Text, "run: go build")
	}
	want := TraceStep{Kind: RefKindCommand, Key: "build", ExpandedTo: "go build", Source: ""}
	if len(res.Trace) != 1 || res.Trace[0] != want {
		t.Fatalf("Trace = %+v, want [%+v]", res.Trace, want)
	}
}

func TestResolveNestedVariableReference(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<b> and more"}
	lookup.variables["b"] = VariableRecord{Name: "b", Value: "bee"}
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("<a>", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "bee and more" {
		t.Fatalf("Text = %q, want %q", res.Text, "bee and more")
	}
	// Inner variable resolves first (depth-first) then outer records its own step.
	if len(res.Trace) != 2 {
		t.Fatalf("Trace = %+v, want 2 steps", res.Trace)
	}
	if res.Trace[0].Key != "b" || res.Trace[1].Key != "a" {
		t.Fatalf("Trace order = %+v, want b then a", res.Trace)
	}
}

func TestResolveCommandReferencingVariable(t *testing.T) {
	lookup := newFakeLookup()
	lookup.commands["greet"] = CommandRecord{Alias: "greet", Template: "echo <name>"}
	lookup.variables["name"] = VariableRecord{Name: "name", Value: "world"}
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("<cmd:greet>", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "echo world" {
		t.Fatalf("Text = %q, want %q", res.Text, "echo world")
	}
}

// --- Resolve: unknown refs, strict vs non-strict ---

func TestResolveUnknownVariableNonStrictLeavesTokenLiteral(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("hello <missing>", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "hello missing" {
		t.Fatalf("Text = %q, want %q", res.Text, "hello missing")
	}
	if len(res.Trace) != 0 {
		t.Fatalf("Trace = %+v, want empty (unresolved refs are not traced)", res.Trace)
	}
}

func TestResolveUnknownVariableStrictReturnsError(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, true, 10)

	_, err := r.Resolve("hello <missing>", rootLabel("root"), nil)
	if err == nil {
		t.Fatal("expected error in strict mode for unknown variable")
	}
	if !errors.Is(err, repository.ErrUnKnownName) {
		t.Fatalf("error = %v, want wrapped ErrUnKnownName", err)
	}
}

func TestResolveUnknownCommandNonStrictLeavesTokenLiteral(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("run <cmd:missing>", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "run cmd:missing" {
		t.Fatalf("Text = %q, want %q", res.Text, "run cmd:missing")
	}
}

func TestResolveUnknownCommandStrictReturnsError(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, true, 10)

	_, err := r.Resolve("run <cmd:missing>", rootLabel("root"), nil)
	if err == nil {
		t.Fatal("expected error in strict mode for unknown command")
	}
	if !errors.Is(err, repository.ErrUnknownAlias) {
		t.Fatalf("error = %v, want wrapped ErrUnknownAlias", err)
	}
}

func TestResolveUnexpectedLookupErrorPropagatesRegardlessOfStrict(t *testing.T) {
	boom := errors.New("db exploded")
	for _, strict := range []bool{true, false} {
		t.Run(fmt.Sprintf("strict=%v", strict), func(t *testing.T) {
			lookup := newFakeLookup()
			lookup.varErr["x"] = boom
			r := NewResolver(lookup, strict, 10)

			_, err := r.Resolve("<x>", rootLabel("root"), nil)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, boom) {
				t.Fatalf("error = %v, want wrapped %v", err, boom)
			}
		})
	}
}

// --- Resolve: escaping ---

func TestResolveEscapedAngleBracketsAreLiteral(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve(`a\<b\>c`, rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "a<b>c" {
		t.Fatalf("Text = %q, want %q", res.Text, "a<b>c")
	}
}

func TestResolveEscapedBackslashIsLiteral(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve(`a\\b`, rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != `a\b` {
		t.Fatalf("Text = %q, want %q", res.Text, `a\b`)
	}
}

func TestResolveUnescapedBackslashBeforeOrdinaryCharIsLiteral(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve(`a\zb`, rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != `a\zb` {
		t.Fatalf("Text = %q, want %q", res.Text, `a\zb`)
	}
}

// --- Resolve: cycle detection ---

func TestResolveDirectSelfCycleDetected(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<a>"}
	r := NewResolver(lookup, false, 10)

	_, err := r.Resolve("<a>", rootLabel("root"), nil)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var cycleErr *CycleDetectionError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("error = %v (%T), want *CycleDetectionError", err, err)
	}
}

func TestResolveIndirectCycleDetected(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<b>"}
	lookup.variables["b"] = VariableRecord{Name: "b", Value: "<a>"}
	r := NewResolver(lookup, false, 10)

	_, err := r.Resolve("<a>", rootLabel("root"), nil)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var cycleErr *CycleDetectionError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("error = %v (%T), want *CycleDetectionError", err, err)
	}
	if !strings.Contains(cycleErr.Error(), "var:a") {
		t.Fatalf("cycle error = %v, want path to mention var:a", cycleErr)
	}
}

func TestResolveCommandCycleDetected(t *testing.T) {
	lookup := newFakeLookup()
	lookup.commands["a"] = CommandRecord{Alias: "a", Template: "<cmd:a>"}
	r := NewResolver(lookup, false, 10)

	_, err := r.Resolve("<cmd:a>", rootLabel("root"), nil)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var cycleErr *CycleDetectionError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("error = %v (%T), want *CycleDetectionError", err, err)
	}
}

func TestResolveRepeatedNonCyclicVariableIsNotFlaggedAsCycle(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "x"}
	r := NewResolver(lookup, false, 10)

	res, err := r.Resolve("<a> and <a> again", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "x and x again" {
		t.Fatalf("Text = %q, want %q", res.Text, "x and x again")
	}
}

// --- Resolve: max depth ---

func TestResolveMaxDepthExceededReturnsError(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<b>"}
	lookup.variables["b"] = VariableRecord{Name: "b", Value: "<c>"}
	lookup.variables["c"] = VariableRecord{Name: "c", Value: "<d>"}
	lookup.variables["d"] = VariableRecord{Name: "d", Value: "leaf"}
	r := NewResolver(lookup, false, 2)

	_, err := r.Resolve("<a>", rootLabel("root"), nil)
	if err == nil {
		t.Fatal("expected max depth error")
	}
	if !strings.Contains(err.Error(), "max depth exceeded") {
		t.Fatalf("error = %v, want max depth exceeded", err)
	}
}

func TestResolveWithinMaxDepthSucceeds(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<b>"}
	lookup.variables["b"] = VariableRecord{Name: "b", Value: "leaf"}
	r := NewResolver(lookup, false, 5)

	res, err := r.Resolve("<a>", rootLabel("root"), nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Text != "leaf" {
		t.Fatalf("Text = %q, want %q", res.Text, "leaf")
	}
}

// --- CollectMissingVars ---

func TestCollectMissingVarsNoRefs(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	missing, err := r.CollectMissingVars("plain text", nil)
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %v, want empty", missing)
	}
}

func TestCollectMissingVarsSuppliedByRuntimeVarsIsNotMissing(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	missing, err := r.CollectMissingVars("<name>", map[string]string{"name": "value"})
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %v, want empty", missing)
	}
}

func TestCollectMissingVarsUnknownVariableIsReported(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	missing, err := r.CollectMissingVars("<name> and <other>", nil)
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	if len(missing) != 2 || missing[0] != "name" || missing[1] != "other" {
		t.Fatalf("missing = %v, want [name other]", missing)
	}
}

func TestCollectMissingVarsDeduplicates(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	missing, err := r.CollectMissingVars("<name> and <name> again", nil)
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	if len(missing) != 1 || missing[0] != "name" {
		t.Fatalf("missing = %v, want [name]", missing)
	}
}

func TestCollectMissingVarsFoundVariableRecursesIntoItsValue(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<b>"}
	r := NewResolver(lookup, false, 10)

	missing, err := r.CollectMissingVars("<a>", nil)
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	if len(missing) != 1 || missing[0] != "b" {
		t.Fatalf("missing = %v, want [b]", missing)
	}
}

func TestCollectMissingVarsUnknownCommandIsHardError(t *testing.T) {
	lookup := newFakeLookup()
	r := NewResolver(lookup, false, 10)

	_, err := r.CollectMissingVars("<cmd:missing>", nil)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !errors.Is(err, repository.ErrUnknownAlias) {
		t.Fatalf("error = %v, want wrapped ErrUnknownAlias", err)
	}
}

func TestCollectMissingVarsFoundCommandRecursesIntoItsTemplate(t *testing.T) {
	lookup := newFakeLookup()
	lookup.commands["build"] = CommandRecord{Alias: "build", Template: "<flag>"}
	r := NewResolver(lookup, false, 10)

	missing, err := r.CollectMissingVars("<cmd:build>", nil)
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	if len(missing) != 1 || missing[0] != "flag" {
		t.Fatalf("missing = %v, want [flag]", missing)
	}
}

func TestCollectMissingVarsUnexpectedLookupErrorPropagates(t *testing.T) {
	boom := errors.New("db exploded")
	lookup := newFakeLookup()
	lookup.varErr["x"] = boom
	r := NewResolver(lookup, false, 10)

	_, err := r.CollectMissingVars("<x>", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want wrapped %v", err, boom)
	}
}

func TestCollectMissingVarsAvoidsInfiniteLoopOnSelfReferencingVariable(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<a>"}
	r := NewResolver(lookup, false, 10)

	missing, err := r.CollectMissingVars("<a>", nil)
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %v, want empty (self-reference already seen)", missing)
	}
}

func TestCollectMissingVarsRespectsMaxDepth(t *testing.T) {
	lookup := newFakeLookup()
	lookup.variables["a"] = VariableRecord{Name: "a", Value: "<b>"}
	lookup.variables["b"] = VariableRecord{Name: "b", Value: "<missing>"}
	r := NewResolver(lookup, false, 0)

	missing, err := r.CollectMissingVars("<a>", nil)
	if err != nil {
		t.Fatalf("CollectMissingVars() error = %v", err)
	}
	// depth 0 processes "<a>" itself (finds "a", recurses at depth 1 which exceeds
	// MaxDepth=0 and stops), so "missing" nested inside b's value is never reached.
	if len(missing) != 0 {
		t.Fatalf("missing = %v, want empty because recursion stops at max depth", missing)
	}
}
