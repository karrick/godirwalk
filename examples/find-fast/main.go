/*
 * find-fast
 *
 * Walks a file system hierarchy using this library.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/karrick/godirwalk"
	"github.com/mattn/go-isatty"
)

var (
	NoColor = os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))
)

func main() {
	optRegex := flag.String("regex", "", "Do not print unless full path matches regex.")
	optQuiet := flag.Bool("quiet", false, "Do not print intermediate errors to stderr.")
	flag.Parse()

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

	dirname := "."
	if flag.NArg() > 0 {
		dirname = flag.Arg(0)
	}

	var sb strings.Builder

	err = godirwalk.Walk(dirname, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if nameRE == nil {
				fmt.Printf("%s\n", osPathname)
				return nil
			}
			if NoColor && nameRE.FindString(osPathname) != "" {
				fmt.Printf("%s\n", osPathname)
			} else if matches := nameRE.FindAllStringSubmatchIndex(osPathname, -1); len(matches) > 0 {
				var prev int
				for _, tuple := range matches {
					sb.WriteString(osPathname[prev:tuple[0]])
					sb.WriteString(yellowOnBlack(osPathname[tuple[0]:tuple[1]]))
					prev = tuple[1]
				}
				sb.WriteString(osPathname[prev:])
				fmt.Println(sb.String())
				sb.Reset()
			}
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			if !*optQuiet {
				fmt.Fprintf(os.Stderr, "%s: %s\n", programName, err)
			}
			// For the purposes of this example, a simple SkipNode will suffice,
			// although in reality perhaps additional logic might be called for.
			return godirwalk.SkipNode
		},
		Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", programName, err)
		os.Exit(1)
	}
}

func yellowOnBlack(s string) string {
	return fmt.Sprintf("\033[1;33;40m%s\033[0m", s)
}
