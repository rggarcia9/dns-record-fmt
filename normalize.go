package dnsfmt

import (
	"fmt"
	"strconv"
	"strings"
)

// knownTypes lists the record types this package has value-specific
// handling for. Anything else still gets name/TTL/class normalization,
// just not value cleanup.
var knownTypes = map[string]bool{
	"A": true, "AAAA": true, "CNAME": true, "MX": true,
	"NS": true, "TXT": true, "PTR": true, "SOA": true,
}

// NormalizeName lowercases a hostname and ensures it ends with a trailing
// dot, the fully-qualified form zone files expect.
func NormalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || strings.HasSuffix(name, ".") {
		return name
	}
	return name + "."
}

// NormalizeType uppercases a record type, e.g. "cname" -> "CNAME".
func NormalizeType(t string) string {
	return strings.ToUpper(strings.TrimSpace(t))
}

// NormalizeClass uppercases the record class and defaults to "IN" when
// empty, since that's what an omitted class means in practice.
func NormalizeClass(class string) string {
	class = strings.ToUpper(strings.TrimSpace(class))
	if class == "" {
		return "IN"
	}
	return class
}

// NormalizeTTL parses a TTL into whole seconds. It accepts plain integers
// ("3600") and single-unit short durations ("30m", "1h", "2d"), since both
// forms show up in hand-edited zone files and copy-pasted DNS panels.
func NormalizeTTL(raw string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return 0, fmt.Errorf("dnsfmt: empty TTL")
	}

	if n, err := strconv.Atoi(raw); err == nil {
		if n < 0 {
			return 0, fmt.Errorf("dnsfmt: negative TTL %q", raw)
		}
		return n, nil
	}

	var multiplier int
	switch raw[len(raw)-1] {
	case 's':
		multiplier = 1
	case 'm':
		multiplier = 60
	case 'h':
		multiplier = 3600
	case 'd':
		multiplier = 86400
	default:
		return 0, fmt.Errorf("dnsfmt: unrecognized TTL %q", raw)
	}

	n, err := strconv.Atoi(raw[:len(raw)-1])
	if err != nil || n < 0 {
		return 0, fmt.Errorf("dnsfmt: unrecognized TTL %q", raw)
	}
	return n * multiplier, nil
}

// NormalizeValue cleans up a record's value field based on its type.
// Domain-valued types get the same lowercase-and-dot treatment as
// NormalizeName; TXT values are wrapped in quotes exactly once; A/AAAA
// are lowercased (IPv6 hex can be mixed case); anything else is trimmed
// and passed through as-is.
func NormalizeValue(recordType, value string) string {
	value = strings.TrimSpace(value)

	switch NormalizeType(recordType) {
	case "CNAME", "NS", "PTR":
		return NormalizeName(value)
	case "MX":
		return normalizeMX(value)
	case "TXT":
		return normalizeTXT(value)
	case "SOA":
		return normalizeSOA(value)
	case "A", "AAAA":
		return strings.ToLower(value)
	default:
		return value
	}
}

// normalizeMX expects "priority host" and rewrites it with the host
// normalized, e.g. "10  Mail.Example.com" -> "10 mail.example.com.". If
// the value doesn't match that shape it's returned unchanged rather than
// guessed at.
func normalizeMX(value string) string {
	fields := strings.Fields(value)
	if len(fields) != 2 {
		return value
	}
	priority, host := fields[0], fields[1]
	if _, err := strconv.Atoi(priority); err != nil {
		return value
	}
	return priority + " " + NormalizeName(host)
}

// normalizeTXT ensures the value is wrapped in double quotes exactly
// once. TXT records are free text and different tools disagree about
// whether the quotes belong in the stored value.
func normalizeTXT(value string) string {
	unquoted := value
	if len(unquoted) >= 2 && strings.HasPrefix(unquoted, `"`) && strings.HasSuffix(unquoted, `"`) {
		unquoted = unquoted[1 : len(unquoted)-1]
	}
	return `"` + unquoted + `"`
}

// normalizeSOA expects the seven whitespace-separated SOA fields --
// primary nameserver, responsible-party mailbox, serial, refresh, retry,
// expire, minimum -- and normalizes the two domain names plus the four
// timing fields, which accept the same unit suffixes as a TTL. The serial
// number is left as a plain integer since it's an opaque counter, not a
// duration. If the value doesn't have exactly seven fields, or any of
// them don't parse, it's returned unchanged rather than guessed at.
func normalizeSOA(value string) string {
	fields := strings.Fields(value)
	if len(fields) != 7 {
		return value
	}

	mname, rname, serial := fields[0], fields[1], fields[2]
	if n, err := strconv.Atoi(serial); err != nil || n < 0 {
		return value
	}

	timings := make([]string, len(fields)-3)
	for i, raw := range fields[3:] {
		ttl, err := NormalizeTTL(raw)
		if err != nil {
			return value
		}
		timings[i] = strconv.Itoa(ttl)
	}

	out := []string{NormalizeName(mname), NormalizeName(rname), serial}
	out = append(out, timings...)
	return strings.Join(out, " ")
}

// IsKnownType reports whether NormalizeValue has type-specific handling
// for t. Unknown types still get their name, TTL, and class normalized;
// only the value is left untouched.
func IsKnownType(t string) bool {
	return knownTypes[NormalizeType(t)]
}
