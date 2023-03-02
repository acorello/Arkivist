package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var renameFlag = flag.Bool("rename", false, "execute renaming")
var basePathFlag = flag.String("path", basePathOrPanic(), "folder containing files to clean-up")

type Config struct {
	rename     bool
	baseFolder string
}

func main() {
	config := initConfig()
	cleanup(config)
}

func initConfig() Config {
	flag.Parse()
	config := Config{
		rename:     *renameFlag,
		baseFolder: *basePathFlag,
	}
	return config
}

func basePathOrPanic() string {
	return filepath.Join(homeDirOrPanic(), "Downloads")
}

func homeDirOrPanic() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Failed to get user's home folder to set default base path")
	}
	return home
}

func cleanup(config Config) {
	folder := config.baseFolder
	for _, dirtyFile := range dirtyFiles(folder) {
		dirtyName := dirtyFile.Name()
		fname := cleanFilename(dirtyName)
		if dirtyName == fname {
			fmt.Println("WARNING: filename cleaning failed")
			fmt.Println("\t- ", dirtyName)
			fmt.Println("\t+ ", fname)
			continue
		}
		if hasUnorthodoxRune(fname) {
			fmt.Println("WARNING: filename cleaning left unorthodox runes in the name")
			fmt.Println("\t- ", dirtyName)
			fmt.Println("\t+ ", fname)
			continue
		}
		oldPath := filepath.Join(folder, dirtyName)
		newPath := filepath.Join(folder, fname)
		if config.rename {
			os.Rename(oldPath, newPath)
		}
		fmt.Println("MOVED:")
		fmt.Println("\t-", oldPath)
		fmt.Println("\t+", newPath)
	}
}

func hasUnorthodoxRune(fname string) bool {
	invalid := regexp.MustCompile(`[^ •[:graph:]]`)
	return invalid.MatchString(fname)
}

func cleanFilename(filename string) string {
	// for reference: `(?i)\s*\(\s*(?:https?...)?z-lib\.org\s*\)\s*`
	replacer := strings.NewReplacer(
		"\u00a0", " ",
		" (z-lib.org)", "",
		" (Z-Library)", "",
		"..", "",
	)
	return replacer.Replace(filename)
}

func dirtyFiles(dir string) (dirtyOnes []fs.DirEntry) {
	isDirty := func(f fs.DirEntry) bool {
		name := strings.ToLower(f.Name())
		return f.Type().IsRegular() && strings.Contains(name, "z-lib")
	}
	allFiles := filesOrPanic(dir)
	dirtyOnes = allFiles[:0]
	for _, f := range allFiles {
		if isDirty(f) {
			dirtyOnes = append(dirtyOnes, f)
		}
	}
	return
}

func filesOrPanic(dir string) []fs.DirEntry {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return files
}
