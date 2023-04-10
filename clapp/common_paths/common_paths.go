package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"dev.acorello.it/go/arkivist/sets"
)

// Given two or more unique directories as arguments
// Output the file paths they have in common; empty directories are ignored.
//
// Es:
//
//	DIR1/
//		A/a.txt
//		B/b.txt
//		D/
//	DIR2/
//		A/a.txt
//		C/c.txt
//		D/
//
// Outputs `A/a.txt`
//
// Input directories are converted to absolute paths and normalized before being compared.
//
// The output is the intersection of the set of subpaths of each directory.
func main() {
	dirs := validatedDirs()
	uniqueFiles := sets.New[string]()
	// (->> dirSet (map list-files) (map set) set/intersection)
	for i, d := range dirs.Entries() {
		if i == 0 {
			collectPaths(d, uniqueFiles)
			continue
		}
		if uniqueFiles.IsEmpty() {
			// whenever I end up with an empty set there is no point in continuing
			break
		}
		currentSet := sets.New[string]()
		collectPaths(d, currentSet)
		// from the the second iteration onwards I have to collect the paths I've already seen and carry over only those ones.
		uniqueFiles = uniqueFiles.Intersection(currentSet)
	}
	for f := range uniqueFiles {
		fmt.Println(f)
	}
}

func collectPaths(dirCleanPath string, currentSet sets.Set[string]) {
	// `dirCleanPath` should always be a clean path for the `relPathOrPanic` to work
	filepath.WalkDir(dirCleanPath, func(path string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return nil
		}
		relPath := relPathOrPanic(dirCleanPath, path)
		currentSet.Add(relPath)
		return nil
	})
}

func relPathOrPanic(baseDir string, path string) string {
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		panic(err)
	}
	return relPath
}

func validatedDirs() sets.Set[string] {
	dirs := sets.New[string]()
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
