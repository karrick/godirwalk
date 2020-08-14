/*
 * find-fast
 *
 * Walks a file system hierarchy using this library.
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/karrick/godirwalk"
	"github.com/karrick/golf"
	"github.com/mattn/go-isatty"
)

var NoColor = os.Getenv("TERM") == "dumb" || !(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))

func main() {
	optQuiet := golf.Bool("quiet", false, "Do not print intermediate errors to stderr.")
	optRegex := golf.String("regex", "", "Do not print unless full path matches regex.")
	optSkip := golf.String("skip", "", "Skip and do not descend into entries with this substring in the pathname")
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
		options.Callback = func(osPathname string, de *godirwalk.Dirent) error {
			if *optSkip != "" && strings.Contains(osPathname, *optSkip) {
				if !*optQuiet {
					fmt.Fprintf(os.Stderr, "%s: %s (skipping)\n", programName, osPathname)
				}
				return godirwalk.SkipThis
			}
			_, err := fmt.Println(osPathname)
			return err
		}
	case NoColor:
		// Name pattern was provided, but color not permitted.
		options.Callback = func(osPathname string, _ *godirwalk.Dirent) error {
			if *optSkip != "" && strings.Contains(osPathname, *optSkip) {
				if !*optQuiet {
					fmt.Fprintf(os.Stderr, "%s: %s (skipping)\n", programName, osPathname)
				}
				return godirwalk.SkipThis
			}
			var err error
			if nameRE.FindString(osPathname) != "" {
				_, err = fmt.Println(osPathname)
			}
			return err
		}
	default:
		// Name pattern provided, and color is permitted.
		buf = append(buf, "\033[22m"...) // very first print should set normal intensity

		options.Callback = func(osPathname string, _ *godirwalk.Dirent) error {
			if *optSkip != "" && strings.Contains(osPathname, *optSkip) {
				if !*optQuiet {
					fmt.Fprintf(os.Stderr, "%s: %s (skipping)\n", programName, osPathname)
				}
				return godirwalk.SkipThis
			}
			matches := nameRE.FindAllStringSubmatchIndex(osPathname, -1)
			if len(matches) == 0 {
				return nil // entry does not match pattern
			}

			var prev int
			for _, tuple := range matches {
				buf = append(buf, osPathname[prev:tuple[0]]...)     // print text before match
				buf = append(buf, "\033[1m"...)                     // bold intensity
				buf = append(buf, osPathname[tuple[0]:tuple[1]]...) // print match
				buf = append(buf, "\033[22m"...)                    // normal intensity
				prev = tuple[1]
			}

			buf = append(buf, osPathname[prev:]...)      // print remaining text after final match
			_, err := os.Stdout.Write(append(buf, '\n')) // don't forget newline
			buf = buf[:0]                                // reset buffer for next string
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
