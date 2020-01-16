/*
 * find-fast
 *
 * Walks a file system hierarchy using this library.
 */
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/karrick/godirwalk"
	"github.com/karrick/golf"
	"github.com/mattn/go-isatty"
	"github.com/xo/terminfo"
)

const useOld = false

var NoColor = false && os.Getenv("TERM") == "dumb" || !(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))

func main() {
	optRegex := golf.String("regex", "", "Do not print unless full path matches regex.")
	optQuiet := golf.Bool("quiet", false, "Do not print intermediate errors to stderr.")
	golf.Parse()

	programName, err := os.Executable()
	if err != nil {
		programName = os.Args[0]
	}
	programName = filepath.Base(programName)

	var nameRE *regexp.Regexp
	if *optRegex != "" {
		nameRE, err = regexp.Compile(*optRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: invalid regex pattern: %s\n", programName, err)
			os.Exit(2)
		}
	}

	var buf []byte // only used when color output

	options := &godirwalk.Options{
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			if !*optQuiet {
				fmt.Fprintf(os.Stderr, "%s: %s\n", programName, err)
			}
			return godirwalk.SkipNode
		},
		Unsorted: true,
	}

	switch {
	case nameRE == nil:
		// When no name pattern provided, print everything.
		options.Callback = func(osPathname string, _ *godirwalk.Dirent) error {
			_, err := fmt.Println(osPathname)
			return err
		}
	case NoColor:
		panic("disabled")
		// Name pattern was provided, but color not permitted.
		options.Callback = func(osPathname string, _ *godirwalk.Dirent) error {
			var err error
			if nameRE.FindString(osPathname) != "" {
				_, err = fmt.Println(osPathname)
			}
			return err
		}
	default:
		// Name pattern provided, and color is permitted.
		var ti *terminfo.Terminfo

		if useOld {
			buf = append(buf, "\033[22m"...) // very first print should set normal intensity
		} else {
			// load terminfo
			ti, err = terminfo.LoadFromEnv()
			if err != nil {
				log.Fatal(err)
			}

			// cleanup
			defer func() {
				err := recover()
				termreset(ti)
				if err != nil {
					log.Fatal(err)
				}
			}()

			terminit(ti)
		}

		options.Callback = func(osPathname string, _ *godirwalk.Dirent) error {
			matches := nameRE.FindAllStringSubmatchIndex(osPathname, -1)
			if len(matches) == 0 {
				return nil // entry does not match pattern
			}

			var prev int
			for _, tuple := range matches {
				if useOld {
					buf = append(buf, osPathname[prev:tuple[0]]...)     // print text before match
					buf = append(buf, "\033[1m"...)                     // bold intensity
					buf = append(buf, osPathname[tuple[0]:tuple[1]]...) // print match
					buf = append(buf, "\033[22m"...)                    // normal intensity
				} else {
					buf = append(buf, termnormal(ti, osPathname[prev:tuple[0]])...)       // print text before match in normal mode
					buf = append(buf, termstandout(ti, osPathname[tuple[0]:tuple[1]])...) // print match in standout mode
				}

				prev = tuple[1]
			}

			if useOld {
				buf = append(buf, osPathname[prev:]...) // print remaining text after final match
				buf = append(buf, '\n')
			} else {
				buf = append(buf, termnormal(ti, osPathname[prev:])...) // print remaining text after final match
				// buf = append(buf, termnewline(ti)...)
				buf = append(buf, '\n')
			}

			_, err := os.Stdout.Write(buf)
			buf = buf[:0] // reset buffer for next string
			return err
		}
	}

	dirname := "."
	if golf.NArg() > 0 {
		dirname = golf.Arg(0)
	}

	if err = godirwalk.Walk(dirname, options); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", programName, err)
		os.Exit(1)
	}
}

// terminit initializes the special CA mode on the terminal, and makes the
// cursor invisible.
func terminit(ti *terminfo.Terminfo) {
	buf := new(bytes.Buffer)
	// enter special mode
	// ti.Fprintf(buf, terminfo.EnterCaMode)
	// // clear the screen
	// ti.Fprintf(buf, terminfo.ClearScreen)
	os.Stdout.Write(buf.Bytes())
}

// termreset is the inverse of terminit.
func termreset(ti *terminfo.Terminfo) {
	buf := new(bytes.Buffer)
	// ti.Fprintf(buf, terminfo.ExitCaMode)
	// ti.Fprintf(buf, terminfo.CursorNormal)
	os.Stdout.Write(buf.Bytes())
}

// termputs puts a string at row, col, interpolating v.
func termputs(ti *terminfo.Terminfo, row, col int, s string, v ...interface{}) []byte {
	buf := new(bytes.Buffer)
	ti.Fprintf(buf, terminfo.CursorAddress, row, col)
	fmt.Fprintf(buf, s, v...)
	return buf.Bytes()
}

func termnormal(ti *terminfo.Terminfo, s string, v ...interface{}) []byte {
	buf := new(bytes.Buffer)
	// ti.Fprintf(buf, terminfo.EnterNormalQuality)
	ti.Fprintf(buf, terminfo.ExitStandoutMode)
	fmt.Fprintf(buf, s, v...)
	return buf.Bytes()
}

func termstandout(ti *terminfo.Terminfo, s string, v ...interface{}) []byte {
	buf := new(bytes.Buffer)
	ti.Fprintf(buf, terminfo.EnterStandoutMode)
	// ti.Fprintf(buf, terminfo.EnterBoldMode)
	fmt.Fprintf(buf, s, v...)
	return buf.Bytes()
}

func termnewline(ti *terminfo.Terminfo) []byte {
	buf := new(bytes.Buffer)
	ti.Fprintf(buf, terminfo.Newline)
	return buf.Bytes()
}
