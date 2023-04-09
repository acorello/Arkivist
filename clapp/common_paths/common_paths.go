package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type fileSet map[string]struct{}

func (I fileSet) Add(a string) {
	I[a] = struct{}{}
}

func main() {
	dirs := validatedArgs()
	uniqueFiles := fileSet{}
	first := true
	// (->> dirSet (map list-files) (map set) set/intersection)
	for d := range dirs {
		currentSet := fileSet{}
		filepath.WalkDir(d, func(path string, info fs.DirEntry, err error) error {
			if info.IsDir() {
				return nil
			}
			path = mustGetRelativePath(d, path)
			currentSet.Add(path)
			return nil
		})
		if first {
			uniqueFiles = currentSet
			first = false
			continue
		}
		commonFiles := fileSet{}
		for f := range currentSet {
			_, found := uniqueFiles[f]
			if found {
				commonFiles.Add(f)
			}
		}
		uniqueFiles = commonFiles
	}
	for f := range uniqueFiles {
		fmt.Println(f)
	}
}

func mustGetRelativePath(baseDir string, path string) string {
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		panic(err)
	}
	return relPath
}

func validatedArgs() fileSet {
	dirs := fileSet{}
	for _, d := range os.Args[1:] {
		a, err := filepath.Abs(d)
		// absolute path
		if err != nil {
			log.Fatalf("Failed to resolve absolute path of %s: %s", d, err.Error())
		}
		s, err := os.Stat(a)
		if err != nil {
			log.Fatalf("Failed to get info for %s: %s", a, err.Error())
		}
		if !s.IsDir() {
			log.Fatalf("Not a directory: %s", a)
		}
		dirs.Add(a)
	}
	if len(dirs) < 2 {
		fmt.Fprintln(os.Stderr, "At least two directories expected")
	}
	return dirs
}
