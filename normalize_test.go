package dnsfmt

import "testing"

func TestNormalizeName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"lowercases and adds dot", "Www.Example.COM", "www.example.com."},
		{"already fully qualified", "example.com.", "example.com."},
		{"trims whitespace", "  example.com  ", "example.com."},
		{"empty string stays empty", "", ""},
		{"bare dot stays as is", ".", "."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeName(c.in); got != c.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeType(t *testing.T) {
	cases := []struct{ in, want string }{
		{"cname", "CNAME"},
		{"  a  ", "A"},
		{"MX", "MX"},
	}
	for _, c := range cases {
		if got := NormalizeType(c.in); got != c.want {
			t.Errorf("NormalizeType(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeClass(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "IN"},
		{"  ", "IN"},
		{"in", "IN"},
		{"CH", "CH"},
		{"hs", "HS"},
	}
	for _, c := range cases {
		if got := NormalizeClass(c.in); got != c.want {
			t.Errorf("NormalizeClass(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeTTL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"plain seconds", "3600", 3600, false},
		{"seconds unit", "45s", 45, false},
		{"minutes unit", "30m", 1800, false},
		{"hours unit", "1h", 3600, false},
		{"days unit", "2d", 172800, false},
		{"uppercase unit", "2D", 172800, false},
		{"whitespace trimmed", "  1h  ", 3600, false},
		{"empty is an error", "", 0, true},
		{"negative plain is an error", "-5", 0, true},
		{"negative with unit is an error", "-5m", 0, true},
		{"unrecognized unit is an error", "5x", 0, true},
		{"garbage is an error", "abc", 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NormalizeTTL(c.in)
			if c.wantErr {
				if err == nil {
					t.Errorf("NormalizeTTL(%q) = %d, nil, want error", c.in, got)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeTTL(%q) returned unexpected error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("NormalizeTTL(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeValue(t *testing.T) {
	cases := []struct {
		name       string
		recordType string
		in         string
		want       string
	}{
		{"CNAME lowercases and qualifies", "CNAME", "Origin.Example.NET", "origin.example.net."},
		{"NS lowercases and qualifies", "ns", "NS1.Example.COM.", "ns1.example.com."},
		{"PTR lowercases and qualifies", "PTR", "Host.Example.com", "host.example.com."},
		{"MX normalizes host, keeps priority", "MX", "10  Mail.Example.com", "10 mail.example.com."},
		{"MX with unparseable priority is untouched", "MX", "ten mail.example.com", "ten mail.example.com"},
		{"MX with wrong field count is untouched", "MX", "10 mail.example.com extra", "10 mail.example.com extra"},
		{"TXT gets quoted", "TXT", "hello world", `"hello world"`},
		{"TXT already quoted stays quoted once", "TXT", `"hello world"`, `"hello world"`},
		{"A is lowercased", "A", "192.0.2.1", "192.0.2.1"},
		{"AAAA is lowercased", "AAAA", "2001:DB8::1", "2001:db8::1"},
		{"unknown type passes through trimmed", "SRV", "  0 5 5060 sip.example.com  ", "0 5 5060 sip.example.com"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeValue(c.recordType, c.in); got != c.want {
				t.Errorf("NormalizeValue(%q, %q) = %q, want %q", c.recordType, c.in, got, c.want)
			}
		})
	}
}

func TestIsKnownType(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"a", true},
		{"AAAA", true},
		{"cname", true},
		{"MX", true},
		{"ns", true},
		{"TXT", true},
		{"ptr", true},
		{"SRV", false},
		{"SOA", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsKnownType(c.in); got != c.want {
			t.Errorf("IsKnownType(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
