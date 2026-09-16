package dnsfmt

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseLine parses one whitespace-separated zone-file-style line into a
// Record and normalizes every field. Accepted forms:
//
//	name TYPE value
//	name TTL TYPE value
//	name CLASS TYPE value
//	name TTL CLASS TYPE value
//	name CLASS TTL TYPE value
//
// TTL and class are both optional and may appear in either order, since
// that's what real zone files and copy-pasted DNS panel exports allow.
// The name is qualified against the root and an omitted TTL is left as 0;
// use ParseRecordLine when $ORIGIN or $TTL directives are in effect.
func ParseLine(line string) (Record, error) {
	return ParseRecordLine(line, ".", 0)
}

// ParseRecordLine is ParseLine, but a name without a trailing dot is
// qualified against origin instead of the root, and an omitted TTL falls
// back to defaultTTL instead of 0. This mirrors how a zone file resolves
// records under the most recent $ORIGIN and $TTL directives.
func ParseRecordLine(line, origin string, defaultTTL int) (Record, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return Record{}, fmt.Errorf("dnsfmt: line has too few fields: %q", line)
	}

	name := fields[0]
	rest := fields[1:]

	ttl := defaultTTL
	var class string

	for len(rest) > 2 {
		token := rest[0]
		if n, err := strconv.Atoi(token); err == nil {
			ttl = n
			rest = rest[1:]
			continue
		}
		if isClassToken(token) {
			class = token
			rest = rest[1:]
			continue
		}
		break
	}

	if len(rest) < 2 {
		return Record{}, fmt.Errorf("dnsfmt: missing type or value in: %q", line)
	}

	recordType := rest[0]
	value := strings.Join(rest[1:], " ")

	return Record{
		Name:  QualifyName(name, origin),
		TTL:   ttl,
		Class: NormalizeClass(class),
		Type:  NormalizeType(recordType),
		Value: NormalizeValueWithOrigin(recordType, value, origin),
	}, nil
}

// ParseOrigin parses a "$ORIGIN example.com." directive line and returns
// the normalized origin, fully qualified and lowercased. ok is false if
// the line isn't a $ORIGIN directive.
func ParseOrigin(line string) (origin string, ok bool) {
	fields := strings.Fields(line)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "$ORIGIN") {
		return "", false
	}
	return NormalizeName(fields[1]), true
}

// ParseTTLDirective parses a "$TTL 3600" directive line and returns the
// TTL in seconds. ok is false if the line isn't a $TTL directive; a
// $TTL directive with a malformed value is reported through err.
func ParseTTLDirective(line string) (ttl int, ok bool, err error) {
	fields := strings.Fields(line)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "$TTL") {
		return 0, false, nil
	}
	ttl, err = NormalizeTTL(fields[1])
	return ttl, true, err
}

func isClassToken(s string) bool {
	switch strings.ToUpper(s) {
	case "IN", "CH", "HS":
		return true
	default:
		return false
	}
}
