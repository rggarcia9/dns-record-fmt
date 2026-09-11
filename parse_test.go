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
