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
	fields := strings.Fields(s)
	if fb, bg, ok := strings.Cut(fields[0], "/"); ok {
		c.Fg = fb
		c.Bg = bg
	} else {
		c.Fg = fields[0]
	}
	for _, tok := range fields[1:] {
		switch tok {
		case "bold":
			c.Bold = true
		case "italic":
			c.Italic = true
		case "faint":
			c.Faint = true
		case "underline":
			c.Underline = true
		default:
			return fmt.Errorf("unknown style attribute: %s", tok)
		}
	}
	return nil
}
