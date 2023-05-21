// # mdfw (mdfind wrapper) an easier API for common mdfind searches
// findbooks this words should be all present
//
// mdfind query language:
// `(` ‹query› `)` grouping
// `( ‹p1› || ‹p2› )` either one of the two predicates
// `( ‹p1› && p2 )` either one of the two predicates
// `attributeName = regex [flags]`
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"
)

const searchPathVariable = "mybooks"

var patternFlag = flag.String("name", "", "pattern for kMDItemDispalyName")
var verboseFlag = flag.Bool("verbose", false, "print mdfind command")
var explicitPathFlag = flag.Bool("explicitpath", false, "print explicit path of result (no env vars)")

func main() {
	flag.Parse()
	/* #!/usr/bin/env fish
	# initial draft implementation in fish
	function findbooks -a pattern -d "Seach in \$mybooks given pattern via `mdfind`"
	    set -l pattern (string escape -n $pattern)
	    set -l prefixLen (math (string length $mybooks) + 2)
	    mdfind -onlyin $mybooks "kMDItemDisplayName = '$pattern'c" | string sub -s $prefixLen
	end
	*/
	searchPath := lookupAndValidateDir(searchPathVariable)
	query := buildQuery(*patternFlag)
	mdfind := exec.Command("mdfind", "-onlyin", searchPath, query)
	if *verboseFlag {
		fmt.Printf("> %s %s %q %q\n", mdfind.Args[0], mdfind.Args[1], mdfind.Args[2], mdfind.Args[3])
	}
	rawResult, err := mdfind.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	results := strings.Split(string(rawResult), "\n")
	printResults(searchPath, results)
}

func lookupAndValidateDir(envVariable string) string {
	dirPath, found := os.LookupEnv(envVariable)
	if !found {
		log.Fatalf("Can't find $%s", envVariable)
	}
	if fs.ValidPath(dirPath) {
		log.Fatalf("$%s not has a valid path: %s", envVariable, dirPath)
	}
	return dirPath
}

func buildQuery(pattern string) string {
	quoteEscaper := strings.NewReplacer(`'`, `\'`, `"`, `\"`)
	pattern = quoteEscaper.Replace(pattern)
	query := fmt.Sprintf("kMDItemDisplayName = '%s'c", pattern)
	return query
}

func printResults(searchPath string, results []string) {
	var out strings.Builder
	prefixLen := len(searchPath)
	for _, line := range results {
		if len(line) == 0 {
			continue
		}
		if *explicitPathFlag {
			out.WriteString(line)
		} else {
			out.WriteString("$")
			out.WriteString(searchPathVariable)
			out.WriteString(line[prefixLen:])
		}
		fmt.Println(out.String())
		out.Reset()
	}
}
