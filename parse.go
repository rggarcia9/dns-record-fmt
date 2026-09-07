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
func ParseLine(line string) (Record, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return Record{}, fmt.Errorf("dnsfmt: line has too few fields: %q", line)
	}

	name := fields[0]
	rest := fields[1:]

	var ttl int
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
		Name:  NormalizeName(name),
		TTL:   ttl,
		Class: NormalizeClass(class),
		Type:  NormalizeType(recordType),
		Value: NormalizeValue(recordType, value),
	}, nil
}

func isClassToken(s string) bool {
	switch strings.ToUpper(s) {
	case "IN", "CH", "HS":
		return true
	default:
		return false
	}
}
