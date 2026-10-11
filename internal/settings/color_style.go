package settings

import (
	"fmt"
	"strings"
)

type ColorStyle struct {
	Fg        string
	Bg        string
	Bold      bool
	Italic    bool
	Faint     bool
	Underline bool
}

// MarshalText serializes the ColorStyle into a textual representation including foreground, background, and style attributes.
func (c ColorStyle) MarshalText() ([]byte, error) {
	if c == (ColorStyle{}) {
		return []byte{}, nil
	}
	var b strings.Builder
	b.WriteString(c.Fg)
	if c.Bg != "" {
		b.WriteString("/")
		b.WriteString(c.Bg)
	}
	for _, a := range []struct {
		set bool
		tok string
	}{{c.Bold, "bold"}, {c.Italic, "italic"}, {c.Faint, "faint"}, {c.Underline, "underline"}} {
		if a.set {
			b.WriteString(" ")
			b.WriteString(a.tok)
		}
	}
	return []byte(b.String()), nil
}

func (c *ColorStyle) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	*c = ColorStyle{}
	if s == "" {
		return nil
	}

	attrs := map[string]func(){
		"bold":      func() { c.Bold = true },
		"italic":    func() { c.Italic = true },
		"faint":     func() { c.Faint = true },
		"underline": func() { c.Underline = true },
	}

	colorSet := false
	for _, tok := range strings.Fields(s) {
		if setAttr, ok := attrs[tok]; ok {
			setAttr()
			continue
		}
		if colorSet {
			return fmt.Errorf("unexpected token %q", tok)
		}
		if fg, bg, ok := strings.Cut(tok, "/"); ok {
			c.Fg = fg
			c.Bg = bg
		} else {
			c.Fg = tok
		}
		colorSet = true
	}
	return nil
}
