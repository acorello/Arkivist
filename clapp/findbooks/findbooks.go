// # mdfw (mdfind wrapper) an easier API for common mdfind searches
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

const searchPath = "mybooks"

func main() {
	/* #!/usr/bin/env fish
	# initial draft implementation in fish
	function findbooks -a pattern -d "Seach in \$mybooks given pattern via `mdfind`"
	    set -l pattern (string escape -n $pattern)
	    set -l prefixLen (math (string length $mybooks) + 2)
	    mdfind -onlyin $mybooks "kMDItemDisplayName = '$pattern'c" | string sub -s $prefixLen
	end
	*/
	patternFlag := flag.String("name", "", "pattern for kMDItemDispalyName")
	verboseFlag := flag.Bool("verbose", false, "print mdfind command")
	explicitPathFlag := flag.Bool("explicitpath", false, "print explicit path of result (no env vars)")
	flag.Parse()
	quoteEscaper := strings.NewReplacer(`'`, `\'`, `"`, `\"`)
	pattern := quoteEscaper.Replace(*patternFlag)
	booksPath, found := os.LookupEnv(searchPath)
	if !found {
		log.Fatalf("Can't find $%s", searchPath)
	}
	if fs.ValidPath(booksPath) {
		log.Fatalf("$%s not has a valid path: %s", searchPath, booksPath)
	}
	query := fmt.Sprintf("kMDItemDisplayName = '%s'c", pattern)
	mdfind := exec.Command("mdfind", "-onlyin", booksPath, query)
	if *verboseFlag {
		fmt.Printf("> %s %s %q %q\n", mdfind.Args[0], mdfind.Args[1], mdfind.Args[2], mdfind.Args[3])
	}
	rawResult, err := mdfind.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	result := strings.Split(string(rawResult), "\n")
	prefixLen := len(booksPath)
	var out strings.Builder
	for _, line := range result {
		if len(line) == 0 {
			continue
		}
		if *explicitPathFlag {
			out.WriteString(line)
		} else {
			out.WriteString("$")
			out.WriteString(searchPath)
			out.WriteString(line[prefixLen:])
		}
		fmt.Println(out.String())
		out.Reset()
	}
}
