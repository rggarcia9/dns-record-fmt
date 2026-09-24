package dnsfmt

import (
	"strings"
	"testing"
)

func TestNormalizeZoneFile(t *testing.T) {
	in := strings.Join([]string{
		"$ORIGIN example.com.",
		"$TTL 3600",
		"; a comment",
		"",
		"www A 192.0.2.1",
		"mail 60 IN MX 10 mail",
		"@ NS ns1",
	}, "\n")

	want := []string{
		"$ORIGIN example.com.",
		"$TTL 3600",
		"; a comment",
		"",
		"www.example.com.\t3600\tIN\tA\t192.0.2.1",
		"mail.example.com.\t60\tIN\tMX\t10 mail.example.com.",
		"example.com.\t3600\tIN\tNS\tns1.example.com.",
	}

	got, err := NormalizeZoneFile(strings.NewReader(in))
	if err != nil {
		t.Fatalf("NormalizeZoneFile returned unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("NormalizeZoneFile returned %d lines, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeZoneFileContinuation(t *testing.T) {
	in := "example.com. 3600 IN SOA ns1.example.com. admin.example.com. (\n" +
		"    2024010101 ; serial\n" +
		"    3600       ; refresh\n" +
		"    900        ; retry\n" +
		"    604800     ; expire\n" +
		"    86400 )    ; minimum\n"

	want := "example.com.\t3600\tIN\tSOA\tns1.example.com. admin.example.com. 2024010101 3600 900 604800 86400"

	got, err := NormalizeZoneFile(strings.NewReader(in))
	if err != nil {
		t.Fatalf("NormalizeZoneFile returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("NormalizeZoneFile(...) = %v, want [%q]", got, want)
	}
}

func TestNormalizeZoneFileErrors(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"unparseable record line", "www CNAME\n"},
		{"bad $TTL value", "$TTL abc\n"},
		{"unclosed continuation", "example.com. IN SOA ns1.example.com. admin.example.com. (\n2024010101\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := NormalizeZoneFile(strings.NewReader(c.in)); err == nil {
				t.Errorf("NormalizeZoneFile(%q) returned nil error, want one", c.in)
			}
		})
	}
}
