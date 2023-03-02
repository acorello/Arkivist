package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	folder := baseFolderOrPanic()
	cleanup(folder)
}

func baseFolderOrPanic() string {
	home, found := os.LookupEnv("HOME")
	if !found {
		log.Fatal("HOME variable not found")
	}
	folder := filepath.Join(home, "Downloads")
	return folder
}

func cleanup(folder string) {
	for _, f := range dirtyFiles(folder) {
		fn := cleanFilename(f.Name())
		old := filepath.Join(folder, f.Name())
		new := filepath.Join(folder, fn)
		os.Rename(old, new)
		fmt.Println("-", old)
		fmt.Println("+", new)
	}
}

func cleanFilename(filename string) string {
	fn := strings.ReplaceAll(filename, "\xA0", " ")
	re := regexp.MustCompile(`\s*\(\s*(?:https?___)?z-lib\.org\s*\)\s*`)
	fn = re.ReplaceAllString(fn, "")
	re = regexp.MustCompile(`\.\.+`)
	fn = re.ReplaceAllString(fn, ".")
	return fn
}

func dirtyFiles(p string) []fs.DirEntry {
	var laundryList []fs.DirEntry
	files, err := os.ReadDir(p)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.Type().IsRegular() && strings.Contains(f.Name(), "z-lib") {
			laundryList = append(laundryList, f)
		}
	}
	return laundryList
}
