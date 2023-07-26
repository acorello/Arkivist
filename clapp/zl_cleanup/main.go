package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"dev.acorello.it/go/arkivist/clapp/zl_cleanup/fileset"
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
	doTrashFlag         = flag.Bool("trash", false, "trash successfully moved files")
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
	doTrash                bool
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
	if I.doTrash && !I.doRun {
		errors = append(errors, fmt.Errorf("'trash' makes sense only with 'run'"))
	}
	errors = append(errors, missingDirectoriesErrors(I.destinationDirectories...)...)
	errors = append(errors, missingDirectoriesErrors(I.sourceDirectory)...)
	return
}

func main() {
	config, err := populateConfig()
	if err != nil {
		log.Fatal("Error parsing config", err)
	}

	exitCode, stop := shouldStop(config)
	if stop {
		os.Exit(exitCode)
	}

	linkToCleanPath(config)
}

func shouldStop(config Config) (exitCode int, stop bool) {
	if *justPrintConfigFlag {
		fmt.Printf("%+v\n", config)
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

func populateConfig() (config Config, err error) {
	flag.Parse()
	config = Config{
		destinationDirectories: *destinationsDirectoryFlag,
		onlyPrintFailed:        *onlyFailedFlag,
		quiet:                  *quietFlag,
		doRun:                  *doRunFlag,
		doTrash:                *doTrashFlag,
		sourceDirectory:        *sourceDirectoryFlag,
		summary:                *summaryFlag,
	}
	if len(config.destinationDirectories) == 0 {
		config.destinationDirectories = []string{config.sourceDirectory}
	}
	config.sourceDirectory, err = filepath.Abs(config.sourceDirectory)
	if err != nil {
		return config, errors.Join(errors.New("filepath.Abs(<source-directory>) failed"), err)
	}
	for i, dir := range config.destinationDirectories {
		config.destinationDirectories[i], err = filepath.Abs(dir)
		if err != nil {
			return config, errors.Join(errors.New("filepath.Abs(<destination-directory>) failed"), err)
		}
	}
	return config, nil
}

type Summary struct {
	out strings.Builder
	err strings.Builder
	Config
}

func (I *Summary) fmtSummary(format string, a ...any) {
	if I.quiet {
		return
	}
	I.out.WriteString(fmt.Sprintf(format, a...))
}

func (I *Summary) fmtErr(format string, a ...any) {
	if I.quiet {
		return
	}
	I.err.WriteString(fmt.Sprintf(format, a...))
}

func (I *Summary) Entry(header, dirPath, fileName string) {
	header = color.GreenString(header + ":")
	I.fmtSummary("%s %s\n\t%s\n", header, dirPath, fileName)
}
func (I *Summary) Source(filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.HiGreenString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.GreenString("%s", dirPath)
	I.Entry("SOURCE", dirPath, fileName)
}
func (I *Summary) LinkPreview(oldPath, filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.HiWhiteString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.WhiteString("%s", dirPath)
	I.Entry("LINK??", dirPath, fileName)
}
func (I *Summary) Linked(oldPath, filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.HiCyanString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.CyanString("%s", dirPath)
	I.Entry("LINKED", dirPath, fileName)
}

func (I *Summary) Homonym(oldPath, filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.HiBlueString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.BlueString("%s", dirPath)
	I.Entry("HMONYM", dirPath, fileName)
}

func (I *Summary) Trashing(filePath string) {
	fileName := filepath.Base(filePath)
	fileName = color.HiYellowString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.YellowString("%s", dirPath)
	I.Entry("TRASHING", dirPath, fileName)
}

func (I *Summary) Error(oldPath, filePath string, err error) {
	fileName := filepath.Base(filePath)
	fileName = color.RedString("%s", fileName)
	dirPath := filepath.Dir(filePath)
	dirPath = color.RedString("%s", dirPath)
	errMessage := color.HiRedString(err.Error())
	header := color.RedString("ERROR:")
	I.fmtErr("%s %s\n\t%s\n\t%s\n", header, dirPath, fileName, errMessage)
}

func (I *Summary) Len() int {
	return I.out.Len() + I.err.Len()
}

func (I *Summary) Print() {
	if !I.quiet && I.Len() == 0 {
		I.fmtSummary("Nothing to report\n")
	}
	fmt.Fprint(os.Stdout, I.out.String())
	fmt.Fprint(os.Stderr, I.err.String())
}

func NewSummary(config Config) Summary {
	return Summary{
		Config: config,
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func linkToCleanPath(config Config) {
	report := NewSummary(config)
	sourceDirectory := config.sourceDirectory
	successfullyLinkedFiles := fileset.New()
	for _, dirtyFile := range dirtyFiles(sourceDirectory) {
		dirtyName := dirtyFile.Name()
		cleanName := cleanFilename(dirtyName)
		if hasFailures(&report, dirtyName, cleanName) || config.onlyPrintFailed {
			continue
		}
		oldPath := filepath.Join(sourceDirectory, dirtyName)
		successfullyLinkedFiles.Add(oldPath) //assume ok, remove if err
		for _, destination := range config.destinationDirectories {
			newPath := filepath.Join(destination, cleanName)
			if config.dryRun() {
				if fileExists(newPath) {
					report.Homonym(oldPath, newPath)
				} else {
					report.LinkPreview(oldPath, newPath)
				}
			} else {
				err, gotErr := os.Link(oldPath, newPath).(*os.LinkError)
				switch {
				case !gotErr:
					report.Linked(oldPath, newPath)
				case errors.Is(err, os.ErrExist):
					report.Homonym(oldPath, newPath)
				default:
					report.Error(err.Old, err.New, err.Err)
					successfullyLinkedFiles.Remove(oldPath)
				}
			}
		}
	}
	if config.doTrash {
		tryTrash(successfullyLinkedFiles, &report)
	}
	report.Print()
}

func tryTrash(movedFiles fileset.FileSet, report *Summary) {
	if movedFiles.IsEmpty() {
		return
	}
	var fileNames strings.Builder
	for fileName := range movedFiles {
		report.Trashing(fileName)
		if fileNames.Len() > 0 {
			fileNames.WriteString(", ")
		}
		fileNames.WriteString(fmt.Sprintf(`POSIX file "%s"`, fileName))
	}
	osascript := fmt.Sprintf(`tell application "Finder" to delete {%s}`, fileNames.String())
	cmd := exec.Command("osascript", "-e", osascript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		report.fmtSummary(color.RedString("ERROR TRASHING FILES:\n%s"), out)
	} else {
		report.fmtSummary(color.GreenString("TRASHED OK:\n%s"), out)
	}
}

func hasFailures(s *Summary, dirtyName, fname string) bool {
	if dirtyName == fname {
		s.Error(dirtyName, fname, fmt.Errorf("name hasn't changed"))
		return true
	}
	if substrings := invalidSubstrings(fname); substrings != nil {
		var sb strings.Builder
		for _, sub := range substrings {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%d: %s", sub.position, sub.value))
		}
		s.Error(dirtyName, fname, fmt.Errorf("found offensive runes at [%s]", sb.String()))
		return true
	}
	return false
}

type invalidSubstring struct {
	position int
	value    string
}

func invalidSubstrings(fname string) (res []invalidSubstring) {
	// \p{L} any letter
	// \p{N} any number
	// \p{P} any punctuation
	// \p{Mn} non-spacing marks
	valid := regexp.MustCompile(`[^ •’\p{L}\p{N}\p{P}\p{Mn}]`)
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
