package dnsfmt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// NormalizeZoneFile reads zone-file-style text from r, one record per
// line, and returns its normalized form as a slice of output lines in
// the same shape the dnsfmt command line tool prints: blank lines and
// comments pass through unchanged, $ORIGIN and $TTL directives are
// tracked and echoed back, and every record line is parsed and
// normalized against the most recently seen origin and default TTL.
// Parenthesis continuation is resolved before a line is parsed.
//
// The first line that fails to parse aborts the whole read and returns
// an error identifying its input line number, rather than normalizing
// what it can and silently dropping the rest -- a caller diffing two
// files wants to know the input was malformed, not compare partial
// output.
func NormalizeZoneFile(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)

	var out []string
	origin := "."
	defaultTTL := 0
	var joiner LineJoiner
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line, ok := joiner.Feed(scanner.Text())
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			out = append(out, line)
			continue
		}

		if newOrigin, ok := ParseOrigin(trimmed); ok {
			origin = newOrigin
			out = append(out, line)
			continue
		}

		if newTTL, ok, err := ParseTTLDirective(trimmed); ok {
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			defaultTTL = newTTL
			out = append(out, line)
			continue
		}

		record, err := ParseRecordLine(line, origin, defaultTTL)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		out = append(out, record.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	if joiner.Pending() {
		return nil, fmt.Errorf("reached end of input with an unclosed \"(\"")
	}
	return out, nil
}
