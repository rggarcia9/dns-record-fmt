package dnsfmt

import "testing"

func TestLineJoinerPassesThroughOrdinaryLines(t *testing.T) {
	var j LineJoiner
	cases := []string{
		"",
		"example.com A 192.0.2.1",
		"; a full-line comment",
		"$ORIGIN example.com.",
	}
	for _, in := range cases {
		out, ok := j.Feed(in)
		if !ok {
			t.Fatalf("Feed(%q) = _, false, want true", in)
		}
		if out != in {
			t.Errorf("Feed(%q) = %q, want unchanged", in, out)
		}
		if j.Pending() {
			t.Errorf("Feed(%q) left the joiner pending", in)
		}
	}
}

func TestLineJoinerJoinsContinuation(t *testing.T) {
	var j LineJoiner
	lines := []string{
		"example.com. IN SOA ns1.example.com. admin.example.com. (",
		"    2024010101 ; serial",
		"    3600       ; refresh",
		"    900        ; retry",
		"    604800     ; expire",
		"    86400 )    ; minimum",
	}

	for _, in := range lines[:len(lines)-1] {
		out, ok := j.Feed(in)
		if ok {
			t.Fatalf("Feed(%q) = %q, true, want false while continuation is open", in, out)
		}
		if !j.Pending() {
			t.Errorf("Feed(%q) should leave the joiner pending", in)
		}
	}

	got, ok := j.Feed(lines[len(lines)-1])
	if !ok {
		t.Fatalf("Feed on closing line = _, false, want true")
	}
	if j.Pending() {
		t.Errorf("joiner still pending after closing paren")
	}

	want := "example.com. IN SOA ns1.example.com. admin.example.com. 2024010101 3600 900 604800 86400"
	if got != want {
		t.Errorf("joined line = %q, want %q", got, want)
	}
}

func TestLineJoinerHandlesParensOpeningAndClosingOnSameLine(t *testing.T) {
	var j LineJoiner
	in := "example.com. IN SOA ns1.example.com. admin.example.com. ( 2024010101 3600 900 604800 86400 )"
	out, ok := j.Feed(in)
	if !ok {
		t.Fatalf("Feed(%q) = _, false, want true", in)
	}
	want := "example.com. IN SOA ns1.example.com. admin.example.com. 2024010101 3600 900 604800 86400"
	if out != want {
		t.Errorf("Feed(%q) = %q, want %q", in, out, want)
	}
}

func TestLineJoinerParensInsideQuotesAreLiteral(t *testing.T) {
	var j LineJoiner
	in := `example.com. TXT "look (no continuation here)"`
	out, ok := j.Feed(in)
	if !ok {
		t.Fatalf("Feed(%q) = _, false, want true", in)
	}
	if out != in {
		t.Errorf("Feed(%q) = %q, want unchanged", in, out)
	}
}

func TestLineJoinerIgnoresStrayClosingParen(t *testing.T) {
	var j LineJoiner
	out, ok := j.Feed("example.com. A 192.0.2.1 )")
	if !ok {
		t.Fatalf("Feed with stray ')' = _, false, want true")
	}
	if j.Pending() {
		t.Errorf("stray ')' should not leave the joiner pending")
	}
	_ = out
}

func TestLineJoinerPending(t *testing.T) {
	var j LineJoiner
	if j.Pending() {
		t.Fatalf("zero-value joiner should not be pending")
	}
	if _, ok := j.Feed("example.com. SOA ns1.example.com. admin.example.com. ("); ok {
		t.Fatalf("Feed with unclosed paren returned ok=true")
	}
	if !j.Pending() {
		t.Errorf("joiner should be pending with an unclosed \"(\"")
	}
}
