// Package dnsfmt normalizes messy DNS record text into a consistent,
// canonical form: lowercase fully-qualified names, uppercase types and
// classes, TTLs as plain seconds, and per-type value cleanup.
package dnsfmt

import (
	"strconv"
	"strings"
)

// Record is a single DNS resource record.
type Record struct {
	Name  string // fully-qualified, lowercase, trailing dot
	TTL   int    // seconds; 0 means "not set"
	Class string // e.g. "IN"
	Type  string // e.g. "A", "MX", "TXT"
	Value string // type-specific, normalized
}

// String renders the record in canonical zone-file form:
//
//	name TTL CLASS TYPE VALUE
//
// The TTL field is omitted when it is 0, matching how zone files leave TTL
// out to inherit a default rather than writing a literal zero.
func (r Record) String() string {
	fields := make([]string, 0, 5)
	fields = append(fields, r.Name)
	if r.TTL > 0 {
		fields = append(fields, strconv.Itoa(r.TTL))
	}
	if r.Class != "" {
		fields = append(fields, r.Class)
	}
	fields = append(fields, r.Type, r.Value)
	return strings.Join(fields, "\t")
}
