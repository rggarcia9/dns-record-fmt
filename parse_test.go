package dnsfmt

import "testing"

func TestParseLine(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Record
	}{
		{
			name: "name type value",
			in:   "example.com A 192.0.2.1",
			want: Record{Name: "example.com.", TTL: 0, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name: "name ttl type value",
			in:   "example.com 300 A 192.0.2.1",
			want: Record{Name: "example.com.", TTL: 300, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name: "name class type value",
			in:   "example.com IN A 192.0.2.1",
			want: Record{Name: "example.com.", TTL: 0, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name: "name ttl class type value",
			in:   "example.com 300 IN A 192.0.2.1",
			want: Record{Name: "example.com.", TTL: 300, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name: "name class ttl type value",
			in:   "example.com IN 300 A 192.0.2.1",
			want: Record{Name: "example.com.", TTL: 300, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name: "mixed case and missing dot are normalized",
			in:   "  Www.Example.com   1h  cname   Origin.Example.NET.  ",
			want: Record{Name: "www.example.com.", TTL: 3600, Class: "IN", Type: "CNAME", Value: "origin.example.net."},
		},
		{
			name: "mx value with extra whitespace",
			in:   "example.com IN MX 10 mail.example.com",
			want: Record{Name: "example.com.", TTL: 0, Class: "IN", Type: "MX", Value: "10 mail.example.com."},
		},
		{
			name: "txt value spans multiple fields",
			in:   "example.com TXT hello world",
			want: Record{Name: "example.com.", TTL: 0, Class: "IN", Type: "TXT", Value: `"hello world"`},
		},
		{
			name: "lowercase class token is recognized",
			in:   "example.com in A 192.0.2.1",
			want: Record{Name: "example.com.", TTL: 0, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name: "soa value normalizes hosts and timing fields",
			in:   "example.com 3600 SOA NS1.Example.com. Admin.Example.com. 2024010101 1h 30m 604800 1d",
			want: Record{
				Name:  "example.com.",
				TTL:   3600,
				Class: "IN",
				Type:  "SOA",
				Value: "ns1.example.com. admin.example.com. 2024010101 3600 1800 604800 86400",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseLine(c.in)
			if err != nil {
				t.Fatalf("ParseLine(%q) returned unexpected error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("ParseLine(%q) = %+v, want %+v", c.in, got, c.want)
			}
		})
	}
}

func TestParseLineErrors(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty line", ""},
		{"single field", "example.com"},
		{"two fields", "example.com A"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := ParseLine(c.in); err == nil {
				t.Errorf("ParseLine(%q) returned nil error, want an error", c.in)
			}
		})
	}
}

func TestParseRecordLineWithOrigin(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		origin     string
		defaultTTL int
		want       Record
	}{
		{
			name:       "relative name qualified against origin",
			in:         "www A 192.0.2.1",
			origin:     "example.com.",
			defaultTTL: 0,
			want:       Record{Name: "www.example.com.", TTL: 0, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name:       "at-sign resolves to the origin",
			in:         "@ MX 10 mail",
			origin:     "example.com.",
			defaultTTL: 0,
			want:       Record{Name: "example.com.", TTL: 0, Class: "IN", Type: "MX", Value: "10 mail.example.com."},
		},
		{
			name:       "absolute name ignores the origin",
			in:         "www.other.com. A 192.0.2.1",
			origin:     "example.com.",
			defaultTTL: 0,
			want:       Record{Name: "www.other.com.", TTL: 0, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name:       "missing ttl falls back to defaultTTL",
			in:         "www A 192.0.2.1",
			origin:     "example.com.",
			defaultTTL: 3600,
			want:       Record{Name: "www.example.com.", TTL: 3600, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
		{
			name:       "explicit ttl overrides defaultTTL",
			in:         "www 60 A 192.0.2.1",
			origin:     "example.com.",
			defaultTTL: 3600,
			want:       Record{Name: "www.example.com.", TTL: 60, Class: "IN", Type: "A", Value: "192.0.2.1"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseRecordLine(c.in, c.origin, c.defaultTTL)
			if err != nil {
				t.Fatalf("ParseRecordLine(%q, %q, %d) returned unexpected error: %v", c.in, c.origin, c.defaultTTL, err)
			}
			if got != c.want {
				t.Errorf("ParseRecordLine(%q, %q, %d) = %+v, want %+v", c.in, c.origin, c.defaultTTL, got, c.want)
			}
		})
	}
}

func TestParseOrigin(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		wantOrigin string
		wantOK     bool
	}{
		{"parses origin directive", "$ORIGIN example.com.", "example.com.", true},
		{"lowercase directive keyword", "$origin Example.COM", "example.com.", true},
		{"adds missing trailing dot", "$ORIGIN example.com", "example.com.", true},
		{"not a directive", "example.com A 192.0.2.1", "", false},
		{"wrong field count", "$ORIGIN", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotOrigin, gotOK := ParseOrigin(c.in)
			if gotOK != c.wantOK || gotOrigin != c.wantOrigin {
				t.Errorf("ParseOrigin(%q) = %q, %v, want %q, %v", c.in, gotOrigin, gotOK, c.wantOrigin, c.wantOK)
			}
		})
	}
}

func TestParseTTLDirective(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantTTL int
		wantOK  bool
		wantErr bool
	}{
		{"parses ttl directive", "$TTL 3600", 3600, true, false},
		{"lowercase directive keyword", "$ttl 1h", 3600, true, false},
		{"not a directive", "example.com A 192.0.2.1", 0, false, false},
		{"wrong field count", "$TTL", 0, false, false},
		{"malformed value is an error", "$TTL abc", 0, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotTTL, gotOK, err := ParseTTLDirective(c.in)
			if c.wantErr {
				if err == nil {
					t.Errorf("ParseTTLDirective(%q) returned nil error, want an error", c.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTTLDirective(%q) returned unexpected error: %v", c.in, err)
			}
			if gotOK != c.wantOK || gotTTL != c.wantTTL {
				t.Errorf("ParseTTLDirective(%q) = %d, %v, want %d, %v", c.in, gotTTL, gotOK, c.wantTTL, c.wantOK)
			}
		})
	}
}

func TestParseLineRoundTripsThroughString(t *testing.T) {
	in := "Www.Example.com 1h CNAME Origin.Example.NET"
	want := "www.example.com.\t3600\tIN\tCNAME\torigin.example.net."

	record, err := ParseLine(in)
	if err != nil {
		t.Fatalf("ParseLine(%q) returned unexpected error: %v", in, err)
	}
	if got := record.String(); got != want {
		t.Errorf("record.String() = %q, want %q", got, want)
	}
}
