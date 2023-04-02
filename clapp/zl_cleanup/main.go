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

	"github.com/fatih/color"
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
	doRunFlag           = flag.Bool("run", false, "execute the operation")
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
	doRun                  bool
	sourceDirectory        string
	summary                bool
}

func (I Config) dryRun() bool {
	return !I.doRun
}

func (I Config) Errors() (errors []error) {
	if I.doRun && I.onlyPrintFailed {
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

	linkToCleanPath(config)
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
		doRun:                  *doRunFlag,
		sourceDirectory:        *sourceDirectoryFlag,
		summary:                *summaryFlag,
	}
	if len(config.destinationDirectories) == 0 {
		config.destinationDirectories = []string{config.sourceDirectory}
	}
	return config
}

type summary struct {
	out strings.Builder
	err strings.Builder
	Config
}

func (I *summary) fmtSummary(format string, a ...any) {
	if I.quiet {
		return
	}
	I.out.WriteString(fmt.Sprintf(format, a...))
}

func (I *summary) fmtErr(format string, a ...any) {
	if I.quiet {
		return
	}
	I.err.WriteString(fmt.Sprintf(format, a...))
}

func (I *summary) fmtEntry(header, dirPath, fileName string) {
	header = color.GreenString(header + ":")
	I.fmtSummary("%s %s\n\t%s\n", header, dirPath, fileName)
}
func (I *summary) fmtSource(filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.CyanString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.BlackString("%s", dirPath)
	I.fmtEntry("SOURCE", dirPath, fileName)
}
func (I *summary) fmtLinkPreview(filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.CyanString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.BlackString("%s", dirPath)
	I.fmtEntry("LINK??", dirPath, fileName)
}
func (I *summary) fmtLink(filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.CyanString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.BlackString("%s", dirPath)
	I.fmtEntry("LINKED", dirPath, fileName)
}

func (I *summary) fmtError(filePath string, err error) {
	fileName := filepath.Base(filePath)
	fileName = color.RedString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.RedString("%s", dirPath)
	errMessage := color.HiRedString(err.Error())
	header := color.RedString("ERROR:")
	I.fmtErr("%s %s\n\t%s\n\t%s\n", header, dirPath, fileName, errMessage)
}

func (I *summary) Len() int {
	return I.out.Len() + I.err.Len()
}

func (I *summary) print() {
	if !I.quiet && I.Len() == 0 {
		I.fmtSummary("Nothing to report\n")
	}
	fmt.Fprint(os.Stdout, I.out.String())
	fmt.Fprint(os.Stderr, I.err.String())
}

func newSummary(config Config) summary {
	return summary{
		Config: config,
	}

}

func linkToCleanPath(config Config) {
	s := newSummary(config)
	sourceDirectory := config.sourceDirectory
	for _, dirtyFile := range dirtyFiles(sourceDirectory) {
		dirtyName := dirtyFile.Name()
		cleanName := cleanFilename(dirtyName)
		if hasFailures(&s, dirtyName, cleanName) || config.onlyPrintFailed {
			continue
		}
		oldPath := filepath.Join(sourceDirectory, dirtyName)
		if !config.quiet {
			s.fmtSource(oldPath)
		}
		for _, destination := range config.destinationDirectories {
			newPath := filepath.Join(destination, cleanName)
			if config.dryRun() {
				s.fmtLinkPreview(newPath)
				continue
			}
			err := os.Link(oldPath, newPath)
			if err != nil {
				s.fmtError(newPath, err)
			} else {
				s.fmtLink(newPath)
			}
		}
	}
	s.print()
}

func hasFailures(s *summary, dirtyName, fname string) bool {
	if dirtyName == fname {
		s.fmtError(dirtyName, fmt.Errorf("failed to clean %q", fname))
		return true
	}
	if invalidSubstrings(fname) != nil {
		s.fmtError(dirtyName, fmt.Errorf("found offensive runes %q", fname))
		return true
	}
	return false
}

type invalidSubstring struct {
	position int
	value    string
}

func invalidSubstrings(fname string) (res []invalidSubstring) {
	valid := regexp.MustCompile(`[^ •’\p{Latin}\p{Nd}[:punct:]]`)
	for _, stringIndices := range valid.FindAllStringIndex(fname, -1) {
		from := stringIndices[0]
		ntil := stringIndices[1]
		substring := fname[from:ntil]
		res = append(res, invalidSubstring{from, substring})
	}
	return
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
		"—", "-",
		"⸺", "-",
		"⸻", "-",
		"﹘", "-",
		"–", "-",
		"‒", "-",
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
