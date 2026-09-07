package resolve

import (
	"fmt"
	"strings"
)

// ReadAngleToken parses a substring starting at the provided index to extract an angle bracket token and its inner content.
// It returns a pointer to the trimmed inner content, the raw token string, and the index after the closing '>'.
// Panics if the substring at startIndex does not begin with '<'.
//
// ReadAngleToken panics if s[startIndex] is not '<'. Callers must confirm this themselves before calling.
func ReadAngleToken(runes []rune, startIndex int) (*string, string, int) {
	n := len(runes)

	if runes[startIndex] != '<' {
		panic(fmt.Errorf("expected '<' at index %d", startIndex))
	}

	i := startIndex + 1
	var innerChars []rune

	for i < n {
		ch := runes[i]
		if ch == '\\' {
			if i+1 < n {
				next := runes[i+1]
				if next == '\\' || next == '<' || next == '>' {
					innerChars = append(innerChars, next)
					i += 2
					continue
				}
			}
			innerChars = append(innerChars, '\\')
			i++
			continue
		}
		if ch == '>' {
			rawToken := string(runes[startIndex+1 : i])
			tokenInner := strings.TrimSpace(string(innerChars))
			return &tokenInner, rawToken, i + 1
		}
		innerChars = append(innerChars, ch)
		i++
	}
	return nil, string(runes[startIndex:]), n
}

// ParseKindAndKey extracts the reference kind and key from a token string in the format "prefix:key" or "key".
func ParseKindAndKey(tokenInner string) (RefKind, string) {
	if !strings.Contains(tokenInner, ":") {
		return RefKindVariable, tokenInner
	}

	parts := strings.SplitN(tokenInner, ":", 2)
	prefix := strings.TrimSpace(parts[0])
	key := strings.TrimSpace(parts[1])

	if prefix == "cmd" {
		return RefKindCommand, key
	}
	if prefix == "var" {
		return RefKindVariable, key
	}
	return RefKindVariable, tokenInner
}

type Ref struct {
	Kind RefKind
	Key  string
}

// ExtractReferences scans the provided template string, identifies and extracts all references enclosed in angle brackets.
// It returns a slice of Ref structs, each containing the kind and key derived from a reference token.
func ExtractReferences(template string) []Ref {
	runes := []rune(template)
	var refs []Ref
	i := 0
	n := len(runes)

	for i < n {
		ch := runes[i]
		if ch == '\\' {
			i += 2
			continue
		}
		if ch == '<' {
			tokenInner, _, nextIdx := ReadAngleToken(runes, i)
			if tokenInner != nil && *tokenInner != "" {
				kind, key := ParseKindAndKey(*tokenInner)
				refs = append(refs, Ref{Kind: kind, Key: key})
			}
			i = nextIdx
			continue
		}
		i++
	}
	return refs
}
