// Command dnsfmt reads DNS records, one per line, from stdin and writes
// their normalized form to stdout. Blank lines and lines starting with
// "#" or ";" are passed through unchanged. $ORIGIN and $TTL directives
// are tracked and applied to the records that follow them, matching how
// a real zone file resolves relative names and default TTLs.
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

	for scanner.Scan() {
		line := scanner.Text()
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
	os.Exit(exitCode)
}
