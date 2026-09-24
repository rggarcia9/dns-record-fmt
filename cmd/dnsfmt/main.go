// Command dnsfmt reads DNS records, one per line, from stdin and writes
// their normalized form to stdout. Blank lines and lines starting with
// "#" or ";" are passed through unchanged. $ORIGIN and $TTL directives
// are tracked and applied to the records that follow them, matching how
// a real zone file resolves relative names and default TTLs. A record
// may also span multiple lines using parenthesis continuation.
//
// With -diff, it instead takes two zone file paths, normalizes each,
// and prints the differences between them.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	dnsfmt "github.com/rggarcia9/dns-record-fmt"
)

func main() {
	diff := flag.Bool("diff", false, "compare the normalized form of two zone files")
	flag.Parse()

	if *diff {
		os.Exit(runDiff(flag.Args()))
	}
	os.Exit(runFormat())
}

func runFormat() int {
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
		return 1
	}
	if joiner.Pending() {
		fmt.Fprintln(os.Stderr, "dnsfmt: reached end of input with an unclosed \"(\"")
		exitCode = 1
	}
	return exitCode
}

// runDiff normalizes the two zone files named in args and prints their
// differences. It returns 0 if the normalized files are identical, 1 if
// they differ, and 2 on a usage or read/parse error, matching the exit
// code convention of the Unix diff command.
func runDiff(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "dnsfmt: -diff requires exactly two file arguments")
		return 2
	}

	linesA, err := normalizeZoneFilePath(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "dnsfmt: %s: %v\n", args[0], err)
		return 2
	}
	linesB, err := normalizeZoneFilePath(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "dnsfmt: %s: %v\n", args[1], err)
		return 2
	}

	changed := false
	for _, line := range dnsfmt.DiffLines(linesA, linesB) {
		fmt.Println(line)
		if line[0] != ' ' {
			changed = true
		}
	}
	if changed {
		return 1
	}
	return 0
}

func normalizeZoneFilePath(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return dnsfmt.NormalizeZoneFile(f)
}
