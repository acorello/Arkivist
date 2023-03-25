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

/* TODO: improve the error message if we call the command without source and destination.

> zl_cleanup -summary
invalid directory: blank
invalid directory: blank

*/

type destinations []string

func (me *destinations) String() string {
	if me == nil {
		return ""
	}
	return strings.Join(*me, string(os.PathListSeparator))
}

func (me *destinations) Set(destination string) error {
	destination = strings.TrimSpace(destination)
	if len(destination) > 0 {
		*me = append(*me, destination)
	}
	return nil
}

var (
	destinationsDirectoryFlag *destinations = new(destinations)

	justPrintConfigFlag = flag.Bool("justconfig", false, "print only the final job configuration")
	onlyFailedFlag      = flag.Bool("onlyfailed", false, "print only files that failed cleanup")
	quietFlag           = flag.Bool("quiet", false, "do not print progress to stdout")
	renameFlag          = flag.Bool("rename", false, "execute renaming")
	sourceDirectoryFlag = flag.String("source", "", "directory containing files to clean-up")
	summaryFlag         = flag.Bool("summary", false, "print list of final filenames at the end")
)

func init() {
	const destinationHelpMsg = "directory where you want to place the ranamed file; can be repeated"
	flag.Var(destinationsDirectoryFlag, "destination", destinationHelpMsg)
}

type Config struct {
	destinationDirectories []string
	onlyPrintFailed        bool
	quiet                  bool
	rename                 bool
	sourceDirectory        string
	summary                bool
}

func (I Config) Errors() (errors []error) {
	if I.rename && I.onlyPrintFailed {
		errors = append(errors, fmt.Errorf("either 'rename' or 'onlyFailed' should be requested"))
	}
	errors = append(errors, missingDirectoriesErrors(I.destinationDirectories...)...)
	errors = append(errors, missingDirectoriesErrors(I.sourceDirectory)...)
	return
}

func main() {
	config := populateConfig()

	exitCode, stop := shouldStop(config)
	if stop {
		os.Exit(exitCode)
	}

	cleanup(config)
}

func shouldStop(config Config) (exitCode int, stop bool) {
	if *justPrintConfigFlag {
		fmt.Printf("%#v\n", config)
		stop = true
	}
	errors := config.Errors()
	for _, err := range errors {
		fmt.Println(err)
	}
	if len(errors) > 0 {
		exitCode = 1
		stop = true
	}
	return
}

func missingDirectoriesErrors(directories ...string) (errors []error) {
	invalidErr := func(dir string) (err error) {
		switch {
		case len(strings.TrimSpace(dir)) == 0:
			return fmt.Errorf("invalid directory: blank")
		case !directoryExists(dir):
			return fmt.Errorf("directory does not exists: %q", dir)
		default:
			return nil
		}
	}
	for _, dir := range directories {
		if err := invalidErr(dir); err != nil {
			errors = append(errors, err)
		}
	}
	return
}

func directoryExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Panicf("Error checking directory: %s\n", err)
		}
	}
	return true
}

func populateConfig() Config {
	flag.Parse()
	config := Config{
		destinationDirectories: *destinationsDirectoryFlag,
		onlyPrintFailed:        *onlyFailedFlag,
		quiet:                  *quietFlag,
		rename:                 *renameFlag,
		sourceDirectory:        *sourceDirectoryFlag,
		summary:                *summaryFlag,
	}
	if len(config.destinationDirectories) == 0 {
		config.destinationDirectories = []string{config.sourceDirectory}
	}
	return config
}

func cleanup(config Config) {
	sourceDirectory := config.sourceDirectory
	var summary strings.Builder
	for _, dirtyFile := range dirtyFiles(sourceDirectory) {
		dirtyName := dirtyFile.Name()
		cleanName := cleanFilename(dirtyName)
		if hasFailures(dirtyName, cleanName) {
			continue
		}
		if config.onlyPrintFailed {
			continue
		}
		oldPath := filepath.Join(sourceDirectory, dirtyName)
		for _, destination := range config.destinationDirectories {
			newPath := filepath.Join(destination, cleanName)
			if config.rename {
				os.Rename(oldPath, newPath)
			}
			if !config.quiet {
				fmt.Println("MOVED:")
				fmt.Println("\t-", oldPath)
				fmt.Println("\t+", newPath)
			}
			if config.summary {
				summary.WriteString(filepath.Base(newPath))
				summary.WriteRune('\n')
			}
		}
	}
	if config.summary {
		if summary.Len() == 0 {
			summary.WriteString("No files moved\n")
		}
		fmt.Print(summary.String())
	}
}

func hasFailures(dirtyName string, fname string) bool {
	printErr := func(s ...string) {
		fmt.Fprintln(os.Stderr, s)
	}
	if dirtyName == fname {
		printErr("WARNING: filename cleaning failed")
		printErr("\t- ", dirtyName)
		printErr("\t+ ", fname)
		return true
	}
	if hasUnorthodoxRune(fname) {
		printErr("WARNING: filename cleaning left unorthodox runes in the name")
		printErr("\t- ", dirtyName)
		printErr("\t+ ", fname)
		return true
	}
	return false
}

func hasUnorthodoxRune(fname string) bool {
	invalid := regexp.MustCompile(`[^ •’[:graph:]]`)
	return invalid.MatchString(fname)
}

func cleanFilename(filename string) string {
	// for reference: `(?i)\s*\(\s*(?:https?...)?z-lib\.org\s*\)\s*`
	const space = " "
	replacer := strings.NewReplacer(
		"\u00a0", space,
		"\xa0", space,
		"\t", space,
		" (z-lib.org)", "",
		" (Z-Library)", "",
		"..", "",
	)

	filename = replacer.Replace(filename)
	rexps := []string{`(\s)\s+`, `(\.)\.+`}
	for _, rexp := range rexps {
		multipleSeparators := regexp.MustCompile(rexp)
		filename = multipleSeparators.ReplaceAllString(filename, `$1`)
	}
	return filename
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
