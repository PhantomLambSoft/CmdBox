package shellinit

import "strings"

const (
	StartMark = "# >>> cmdbox shell integration >>>"
	EndMark   = "# <<< cmdbox shell integration <<<"
)

// UpsertBlock inserts or replaces the marked shell integration block in existing content.
// It preserves CRLF line endings when present and returns ErrUnterminatedBlock for an unmatched start marker.
func UpsertBlock(existing, block string) (string, error) {
	eol := "\n"
	if strings.Contains(existing, "\r\n") {
		eol = "\r\n"
	}

	body := strings.ReplaceAll(block, "\r\n", "\n")
	body = strings.TrimRight(body, "\n \t")
	marked := StartMark + "\n" + body + "\n" + EndMark + "\n"
	if eol != "\n" {
		marked = strings.ReplaceAll(marked, "\n", eol)
	}

	if start := strings.Index(existing, StartMark); start >= 0 {
		endRel := strings.Index(existing[start:], EndMark)
		if endRel < 0 {
			return "", ErrUnterminatedBlock
		}
		end := start + endRel + len(EndMark)
		switch {
		case strings.HasPrefix(existing[end:], "\r\n"):
			end += 2
		case strings.HasPrefix(existing[end:], "\n"):
			end++
		}
		return existing[:start] + marked + existing[end:], nil
	}

	var b strings.Builder
	b.WriteString(existing)
	if existing != "" {
		if !strings.HasSuffix(existing, "\n") {
			b.WriteString(eol)
		}
		if !strings.HasSuffix(existing, eol+eol) {
			b.WriteString(eol) // blank line between user content and our block
		}
	}
	b.WriteString(marked)
	return b.String(), nil
}
