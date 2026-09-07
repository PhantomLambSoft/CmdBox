package resolve

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

type Resolver struct {
	Lookup   Lookup
	Strict   bool
	MaxDepth int
}

func NewResolver(lookup Lookup, strict bool, maxDepth int) Resolver {
	return Resolver{
		Lookup:   lookup,
		Strict:   strict,
		MaxDepth: maxDepth,
	}
}

// CollectMissingVars identifies placeholders in the template that are missing in the provided runtimeVars map and returns them.
func (r *Resolver) CollectMissingVars(template string, runtimeVars map[string]string) ([]string, error) {
	var missing []string
	seen := make(map[string]struct{})
	err := r.collectMissingInner(template, runtimeVars, &missing, seen, 0)
	if err != nil {
		return nil, fmt.Errorf("collecting missing vars: %w", err)
	}
	return missing, nil
}

// Resolve processes a template by recursively resolving variable and command references into a final resolved result.
func (r *Resolver) Resolve(template string, rootLabel *string, runtimeVars map[string]string) (Result, error) {
	var trace []TraceStep
	stack := []string{*rootLabel}
	text, err := r.resolveInner(template, stack, 0, &trace, runtimeVars)
	if err != nil {
		return Result{}, err
	}
	return Result{Text: text, Trace: trace}, nil
}

// resolveInner recursively processes a template to replace variables or commands with resolved values.
// It handles escape sequences, angle bracket tokens, and detects cyclic references up to a maximum depth.
func (r *Resolver) resolveInner(
	template string,
	stack []string,
	depth int,
	trace *[]TraceStep,
	runtimeVars map[string]string,
) (string, error) {
	if depth > r.MaxDepth {
		return "", fmt.Errorf("max depth exceeded")
	}

	var out []string
	i := 0
	runes := []rune(template)
	runeLength := len(runes)

	for i < runeLength {
		ch := runes[i]
		if ch == '\\' {
			if i+1 < runeLength {
				next := runes[i+1]
				if next == '\\' || next == '<' || next == '>' {
					out = append(out, string(next))
					i += 2
					continue
				}
			}
		}

		if ch == '<' {
			tokenInner, rawToken, nextIdx := ReadAngleToken(runes, i)
			if tokenInner == nil {
				out = append(out, "<")
				i++
				continue
			}
			replacement, err := r.expandAngleTokens(
				tokenInner,
				rawToken,
				stack,
				depth,
				trace,
				runtimeVars,
			)
			if err != nil {
				return "", fmt.Errorf("expanding angle tokens: %w", err)
			}
			out = append(out, replacement)
			i = nextIdx
			continue
		}
		out = append(out, string(ch))
		i++
	}
	return strings.Join(out, ""), nil
}

// collectMissingInner identifies unresolved variables or commands in the template and appends them to the missing list.
func (r *Resolver) collectMissingInner(
	template string,
	runtimeVars map[string]string,
	missing *[]string,
	seen map[string]struct{}, // set
	depth int,
) error {
	if depth > r.MaxDepth {
		return nil
	}

	i := 0
	runes := []rune(template)
	for i < len(runes) {
		if runes[i] == '\\' {
			i += 2
			continue
		}
		if runes[i] == '<' {
			tokenInner, _, nextIdx := ReadAngleToken(runes, i)
			if tokenInner != nil {
				kind, key := ParseKindAndKey(*tokenInner)
				if kind == RefKindVariable {
					if runtimeVars != nil {
						if _, ok := runtimeVars[key]; ok {
							i = nextIdx
							continue
						}
					}
					rec, err := r.Lookup.GetVariable(key)
					if err != nil {
						if errors.Is(err, repository.ErrUnKnownName) {
							if !slices.Contains(*missing, key) {
								*missing = append(*missing, key)
							}
						} else {
							return fmt.Errorf("getting variable %s: %w", key, err)
						}
					} else {
						if _, ok := seen[rec.Name]; !ok {
							seen[rec.Name] = struct{}{}
							err = r.collectMissingInner(rec.Value, runtimeVars, missing, seen, depth+1)
							if err != nil {
								return err
							}
						}
					}
				} else {
					rec, err := r.Lookup.GetCommand(key)
					if err != nil {
						return fmt.Errorf("getting command %s: %w", key, err)
					}
					if _, ok := seen[rec.Alias]; !ok {
						seen[rec.Alias] = struct{}{}
						err = r.collectMissingInner(rec.Template, runtimeVars, missing, seen, depth+1)
						if err != nil {
							return err
						}
					}
				}
			}
			i = nextIdx
		} else {
			i++
		}
	}
	return nil
}

// expandAngleTokens resolves tokens enclosed in angle brackets by expanding variables or commands within a given context.
// It handles variable lookups, command transformations, cycle detection, and nested token expansion recursively.
func (r *Resolver) expandAngleTokens(
	tokenInner *string,
	rawToken string,
	stack []string,
	depth int,
	trace *[]TraceStep,
	runtimeVars map[string]string,
) (string, error) {
	if tokenInner == nil || *tokenInner == "" {
		return rawToken, nil
	}

	kind, key := ParseKindAndKey(*tokenInner)
	if kind == RefKindVariable {

		if runtimeVars != nil {
			if value, ok := runtimeVars[key]; ok {
				*trace = append(*trace, TraceStep{Kind: kind, Key: key, ExpandedTo: value, Source: "runtime"})
				return value, nil
			}
		}

		rec, err := r.Lookup.GetVariable(key)
		if err != nil {
			if errors.Is(err, repository.ErrUnKnownName) {
				if r.Strict {
					return "", fmt.Errorf("resolving variable %s in strict mode: %w", key, err)
				}
				return rawToken, nil
			}
			return "", fmt.Errorf("resolving variable %s: %w", key, err)
		}

		label := "var:" + rec.Name
		err = r.checkCycle(label, stack)
		if err != nil {
			return "", err
		}

		stack = append(stack, label)

		expanded, err := r.resolveInner(rec.Value, stack, depth+1, trace, map[string]string{})
		if err != nil {
			return "", fmt.Errorf("resolving variable %s: %w", key, err)
		}

		*trace = append(*trace, TraceStep{Kind: kind, Key: rec.Name, ExpandedTo: expanded, Source: "stored"})
		return expanded, nil
	}

	// Fallback to command lookup
	rec, err := r.Lookup.GetCommand(key)
	if err != nil {
		if errors.Is(err, repository.ErrUnknownAlias) {
			if r.Strict {
				return "", fmt.Errorf("resolving command %s in strict mode: %w", key, err)
			}
			return rawToken, nil
		}
		return "", fmt.Errorf("resolving command %s: %w", key, err)
	}

	label := "cmd:" + rec.Alias
	err = r.checkCycle(label, stack)
	if err != nil {
		return "", err
	}
	stack = append(stack, label)

	expanded, err := r.resolveInner(rec.Template, stack, depth+1, trace, map[string]string{})
	if err != nil {
		return "", fmt.Errorf("resolving command %s: %w", key, err)
	}
	*trace = append(*trace, TraceStep{Kind: RefKindCommand, Key: rec.Alias, ExpandedTo: expanded, Source: ""})

	return expanded, nil
}

// checkCycle detects cycles in the stack by checking if the nextLabel already exists and returns an error if a
// cycle is found.
func (r *Resolver) checkCycle(nextLabel string, stack []string) error {
	for i, label := range stack {
		if label == nextLabel {
			errStack := slices.Clone(stack[i:])
			errStack = append(errStack, nextLabel)
			return &CycleDetectionError{path: errStack}
		}
	}
	return nil
}
