package resolve

import "testing"

// --- ReadAngleToken ---

func TestReadAngleTokenPanicsWhenNotAngleBracket(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when runes[startIndex] is not '<'")
		}
	}()
	ReadAngleToken([]rune("abc"), 0)
}

func TestReadAngleTokenBasic(t *testing.T) {
	runes := []rune("<var:a>")
	inner, raw, next := ReadAngleToken(runes, 0)
	if inner == nil || *inner != "var:a" {
		t.Fatalf("inner = %v, want %q", inner, "var:a")
	}
	if raw != "var:a" {
		t.Fatalf("raw = %q, want %q", raw, "var:a")
	}
	if next != len(runes) {
		t.Fatalf("next = %d, want %d", next, len(runes))
	}
}

func TestReadAngleTokenTrimsWhitespace(t *testing.T) {
	runes := []rune("< var:a >")
	inner, _, _ := ReadAngleToken(runes, 0)
	if inner == nil || *inner != "var:a" {
		t.Fatalf("inner = %v, want %q", inner, "var:a")
	}
}

func TestReadAngleTokenEmptyInner(t *testing.T) {
	runes := []rune("<>")
	inner, raw, next := ReadAngleToken(runes, 0)
	if inner == nil || *inner != "" {
		t.Fatalf("inner = %v, want empty string", inner)
	}
	if raw != "" {
		t.Fatalf("raw = %q, want empty", raw)
	}
	if next != 2 {
		t.Fatalf("next = %d, want 2", next)
	}
}

func TestReadAngleTokenUnterminated(t *testing.T) {
	runes := []rune("<abc")
	inner, raw, next := ReadAngleToken(runes, 0)
	if inner != nil {
		t.Fatalf("inner = %v, want nil", inner)
	}
	if raw != "<abc" {
		t.Fatalf("raw = %q, want %q", raw, "<abc")
	}
	if next != len(runes) {
		t.Fatalf("next = %d, want %d", next, len(runes))
	}
}

func TestReadAngleTokenEscapedSpecialChars(t *testing.T) {
	runes := []rune(`<a\>b>`)
	inner, raw, next := ReadAngleToken(runes, 0)
	if inner == nil || *inner != "a>b" {
		t.Fatalf("inner = %v, want %q", inner, "a>b")
	}
	if raw != `a\>b` {
		t.Fatalf("raw = %q, want %q", raw, `a\>b`)
	}
	if next != len(runes) {
		t.Fatalf("next = %d, want %d", next, len(runes))
	}
}

func TestReadAngleTokenEscapedBackslashAndAngle(t *testing.T) {
	runes := []rune(`<a\\b\<c>`)
	inner, _, _ := ReadAngleToken(runes, 0)
	if inner == nil || *inner != `a\b<c` {
		t.Fatalf("inner = %v, want %q", inner, `a\b<c`)
	}
}

func TestReadAngleTokenBackslashBeforeNonSpecialCharIsLiteral(t *testing.T) {
	runes := []rune(`<a\zb>`)
	inner, _, _ := ReadAngleToken(runes, 0)
	if inner == nil || *inner != `a\zb` {
		t.Fatalf("inner = %v, want %q", inner, `a\zb`)
	}
}

func TestReadAngleTokenTrailingBackslashUnterminated(t *testing.T) {
	runes := []rune(`<abc\`)
	inner, raw, next := ReadAngleToken(runes, 0)
	if inner != nil {
		t.Fatalf("inner = %v, want nil", inner)
	}
	if raw != `<abc\` {
		t.Fatalf("raw = %q, want %q", raw, `<abc\`)
	}
	if next != len(runes) {
		t.Fatalf("next = %d, want %d", next, len(runes))
	}
}

func TestReadAngleTokenNestedAngleBracketTreatedLiterally(t *testing.T) {
	runes := []rune("<<inner>>")
	inner, raw, next := ReadAngleToken(runes, 0)
	if inner == nil || *inner != "<inner" {
		t.Fatalf("inner = %v, want %q", inner, "<inner")
	}
	if raw != "<inner" {
		t.Fatalf("raw = %q, want %q", raw, "<inner")
	}
	if next != 8 {
		t.Fatalf("next = %d, want 8", next)
	}
}

func TestReadAngleTokenFromNonZeroStart(t *testing.T) {
	runes := []rune("xx<var:a>yy")
	inner, raw, next := ReadAngleToken(runes, 2)
	if inner == nil || *inner != "var:a" {
		t.Fatalf("inner = %v, want %q", inner, "var:a")
	}
	if raw != "var:a" {
		t.Fatalf("raw = %q, want %q", raw, "var:a")
	}
	if next != 9 {
		t.Fatalf("next = %d, want 9", next)
	}
}

// --- ParseKindAndKey ---

func TestParseKindAndKeyNoColonIsVariable(t *testing.T) {
	kind, key := ParseKindAndKey("foo")
	if kind != RefKindVariable || key != "foo" {
		t.Fatalf("got (%v, %q), want (%v, %q)", kind, key, RefKindVariable, "foo")
	}
}

func TestParseKindAndKeyCmdPrefix(t *testing.T) {
	kind, key := ParseKindAndKey("cmd:build")
	if kind != RefKindCommand || key != "build" {
		t.Fatalf("got (%v, %q), want (%v, %q)", kind, key, RefKindCommand, "build")
	}
}

func TestParseKindAndKeyVarPrefix(t *testing.T) {
	kind, key := ParseKindAndKey("var:name")
	if kind != RefKindVariable || key != "name" {
		t.Fatalf("got (%v, %q), want (%v, %q)", kind, key, RefKindVariable, "name")
	}
}

func TestParseKindAndKeyUnknownPrefixFallsBackToWholeToken(t *testing.T) {
	kind, key := ParseKindAndKey("other:foo")
	if kind != RefKindVariable || key != "other:foo" {
		t.Fatalf("got (%v, %q), want (%v, %q)", kind, key, RefKindVariable, "other:foo")
	}
}

func TestParseKindAndKeyTrimsWhitespaceAroundPrefixAndKey(t *testing.T) {
	kind, key := ParseKindAndKey(" cmd : build ")
	if kind != RefKindCommand || key != "build" {
		t.Fatalf("got (%v, %q), want (%v, %q)", kind, key, RefKindCommand, "build")
	}
}

func TestParseKindAndKeyMultipleColonsSplitsOnFirst(t *testing.T) {
	kind, key := ParseKindAndKey("cmd:foo:bar")
	if kind != RefKindCommand || key != "foo:bar" {
		t.Fatalf("got (%v, %q), want (%v, %q)", kind, key, RefKindCommand, "foo:bar")
	}
}

// --- ExtractReferences ---

func TestExtractReferencesNoRefs(t *testing.T) {
	refs := ExtractReferences("plain text with no refs")
	if len(refs) != 0 {
		t.Fatalf("refs = %v, want empty", refs)
	}
}

func TestExtractReferencesSingleVariable(t *testing.T) {
	refs := ExtractReferences("hello <name>")
	want := []Ref{{Kind: RefKindVariable, Key: "name"}}
	if len(refs) != 1 || refs[0] != want[0] {
		t.Fatalf("refs = %v, want %v", refs, want)
	}
}

func TestExtractReferencesSingleCommand(t *testing.T) {
	refs := ExtractReferences("<cmd:build>")
	want := []Ref{{Kind: RefKindCommand, Key: "build"}}
	if len(refs) != 1 || refs[0] != want[0] {
		t.Fatalf("refs = %v, want %v", refs, want)
	}
}

func TestExtractReferencesMultipleMixed(t *testing.T) {
	refs := ExtractReferences("hello <name>, run <cmd:build> then <var:other>")
	want := []Ref{
		{Kind: RefKindVariable, Key: "name"},
		{Kind: RefKindCommand, Key: "build"},
		{Kind: RefKindVariable, Key: "other"},
	}
	if len(refs) != len(want) {
		t.Fatalf("refs = %v, want %v", refs, want)
	}
	for i := range want {
		if refs[i] != want[i] {
			t.Fatalf("refs[%d] = %v, want %v", i, refs[i], want[i])
		}
	}
}

func TestExtractReferencesEscapedAngleBracketIsNotAToken(t *testing.T) {
	refs := ExtractReferences(`\<not a token>`)
	if len(refs) != 0 {
		t.Fatalf("refs = %v, want empty", refs)
	}
}

func TestExtractReferencesEmptyTokenIsSkipped(t *testing.T) {
	refs := ExtractReferences("<>")
	if len(refs) != 0 {
		t.Fatalf("refs = %v, want empty", refs)
	}
}

func TestExtractReferencesWhitespaceOnlyTokenIsSkipped(t *testing.T) {
	refs := ExtractReferences("< >")
	if len(refs) != 0 {
		t.Fatalf("refs = %v, want empty", refs)
	}
}

func TestExtractReferencesUnterminatedTokenIsIgnored(t *testing.T) {
	refs := ExtractReferences("text <abc")
	if len(refs) != 0 {
		t.Fatalf("refs = %v, want empty", refs)
	}
}
