# dns-record-fmt

Every source of DNS records writes them a little differently: mixed-case
hostnames, missing trailing dots, TTLs written as "1h" instead of 3600,
TXT values with or without quotes, MX values with extra whitespace between
the priority and the host. Diffing or comparing records pulled from two
different places (a zone file export, a registrar's UI, an API response)
is miserable until everything is in the same shape.

`dns-record-fmt` normalizes that mess into one canonical form:

```
name    TTL    CLASS    TYPE    VALUE
```

- names are lowercased and made fully-qualified (trailing dot)
- types and classes are uppercased, class defaults to `IN`
- TTLs are converted to plain seconds (`1h` -> `3600`)
- values are cleaned up per record type: domain names in CNAME/NS/PTR/MX
  get the same lowercase-and-dot treatment, TXT values are quoted exactly
  once, A/AAAA addresses are lowercased, and SOA records get their two
  hostnames qualified and their refresh/retry/expire/minimum fields
  converted to plain seconds (the serial number is left alone)

## Library usage

```go
package main

import (
	"fmt"

	dnsfmt "github.com/rggarcia9/dns-record-fmt"
)

func main() {
	r, err := dnsfmt.ParseLine("  Www.Example.com   1h  cname   Origin.Example.NET.  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// www.example.com.	3600	IN	CNAME	origin.example.net.
}
```

Every exported function takes plain values in and returns plain values
out — no I/O, no globals, no hidden state — so each one can be tested with
a table of inputs and expected outputs:

```go
ttl, err := dnsfmt.NormalizeTTL("2d")
// ttl == 172800, err == nil

name := dnsfmt.NormalizeName("Example.COM")
// name == "example.com."
```

## Command-line usage

`cmd/dnsfmt` reads records from stdin, one per line, and writes the
normalized form to stdout:

```
$ printf 'www.example.com 3600 CNAME origin.example.net\nexample.com IN MX 10 mail.example.com\n' | go run ./cmd/dnsfmt
www.example.com.	3600	IN	CNAME	origin.example.net.
example.com.	IN	MX	10 mail.example.com.
```

Lines that are blank, or start with `#` or `;`, are passed through
unchanged so comments in a copy-pasted zone file survive.

`$ORIGIN` and `$TTL` directives are tracked as the file is read and
applied to the records that follow, the same way a real zone file does:

```
$ printf '$ORIGIN example.com.\n$TTL 3600\nwww A 192.0.2.1\nmail 60 IN MX 10 mail\n@ NS ns1\n' | go run ./cmd/dnsfmt
$ORIGIN example.com.
$TTL 3600
www.example.com.	3600	IN	A	192.0.2.1
mail.example.com.	60	IN	MX	10 mail.example.com.
example.com.	3600	IN	NS	ns1.example.com.
```

A name of `@` refers to the origin itself, and any name without a
trailing dot is qualified against the most recent `$ORIGIN` rather than
the zone root.

## Status

Early. The parser covers the common single-line record shapes (A, AAAA,
CNAME, MX, NS, TXT, PTR, SOA) and `$ORIGIN` / `$TTL` directives, but not
multi-line records with parenthesis continuation yet.

## License

MIT, see [LICENSE](LICENSE).
