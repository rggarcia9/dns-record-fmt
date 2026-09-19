// Command dnsfmt reads DNS records, one per line, from stdin and writes
// their normalized form to stdout. Blank lines and lines starting with
// "#" or ";" are passed through unchanged. $ORIGIN and $TTL directives
// are tracked and applied to the records that follow them, matching how
// a real zone file resolves relative names and default TTLs. A record
// may also span multiple lines using parenthesis continuation.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	dnsfmt "github.com/rggarcia9/dns-record-fmt"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	exitCode := 0

	origin := "."
	defaultTTL := 0
	var joiner dnsfmt.LineJoiner

	for scanner.Scan() {
		line, ok := joiner.Feed(scanner.Text())
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			fmt.Println(line)
			continue
		}

		if newOrigin, ok := dnsfmt.ParseOrigin(trimmed); ok {
			origin = newOrigin
			fmt.Println(line)
			continue
		}

		if newTTL, ok, err := dnsfmt.ParseTTLDirective(trimmed); ok {
			if err != nil {
				fmt.Fprintf(os.Stderr, "dnsfmt: %v\n", err)
				exitCode = 1
				continue
			}
			defaultTTL = newTTL
			fmt.Println(line)
			continue
		}

		record, err := dnsfmt.ParseRecordLine(line, origin, defaultTTL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dnsfmt: %v\n", err)
			exitCode = 1
			continue
		}
		fmt.Println(record.String())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "dnsfmt: reading input: %v\n", err)
		os.Exit(1)
	}
	if joiner.Pending() {
		fmt.Fprintln(os.Stderr, "dnsfmt: reached end of input with an unclosed \"(\"")
		exitCode = 1
	}
	os.Exit(exitCode)
}
