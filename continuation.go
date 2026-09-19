package dnsfmt

import "strings"

// LineJoiner resolves the parenthesis continuation zone files use to
// spread one record across several physical lines, e.g.:
//
//	example.com. IN SOA ns1.example.com. admin.example.com. (
//	    2024010101 ; serial
//	    3600       ; refresh
//	    900        ; retry
//	    604800     ; expire
//	    86400 )    ; minimum
//
// The zero value is ready to use.
type LineJoiner struct {
	buf   strings.Builder
	depth int
}

// Feed adds one physical line. When the line isn't part of an open
// continuation, it's returned unchanged with ok set to true so blank
// lines, directives, and full-line comments pass straight through.
// Once an unmatched "(" opens a continuation, Feed buffers lines --
// stripping comments and the parentheses themselves -- until the
// matching ")" closes it, then returns the joined logical line. A ")"
// with no open "(" is ignored rather than treated as an error, since a
// stray one shouldn't wedge every line after it into a continuation.
func (j *LineJoiner) Feed(line string) (out string, ok bool) {
	uncommented := stripComment(line)
	balance := parenBalance(uncommented)

	if j.depth == 0 && balance == 0 {
		return line, true
	}

	if content := strings.TrimSpace(stripParens(uncommented)); content != "" {
		if j.buf.Len() > 0 {
			j.buf.WriteByte(' ')
		}
		j.buf.WriteString(content)
	}

	j.depth += balance
	if j.depth <= 0 {
		j.depth = 0
		out = j.buf.String()
		j.buf.Reset()
		return out, true
	}
	return "", false
}

// Pending reports whether Feed is mid-continuation, waiting on a closing
// paren that never arrived. A caller reaching end of input with this true
// has a malformed record.
func (j *LineJoiner) Pending() bool {
	return j.depth > 0
}

// stripComment removes a ";" comment and everything after it on the
// line, unless the ";" is inside a double-quoted string.
func stripComment(line string) string {
	inQuotes := false
	for i, r := range line {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case ';':
			if !inQuotes {
				return line[:i]
			}
		}
	}
	return line
}

// parenBalance counts unquoted "(" as +1 and ")" as -1.
func parenBalance(line string) int {
	balance := 0
	inQuotes := false
	for _, r := range line {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case '(':
			if !inQuotes {
				balance++
			}
		case ')':
			if !inQuotes {
				balance--
			}
		}
	}
	return balance
}

// stripParens blanks out unquoted "(" and ")" so the surrounding fields
// can be joined without the continuation markers ending up in the value.
func stripParens(line string) string {
	inQuotes := false
	var b strings.Builder
	for _, r := range line {
		switch r {
		case '"':
			inQuotes = !inQuotes
			b.WriteRune(r)
		case '(', ')':
			if inQuotes {
				b.WriteRune(r)
			} else {
				b.WriteRune(' ')
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
