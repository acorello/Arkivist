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
		out.WriteString("$")
		out.WriteString(searchPath)
		out.WriteString(line[prefixLen:])
		fmt.Println(out.String())
		out.Reset()
	}
}
