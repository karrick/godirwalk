package main

import (
	"fmt"
	"os"

	"github.com/karrick/godirwalk"
)

func main() {
	dirname := "."
	if len(os.Args) > 1 {
		dirname = os.Args[1]
	}
	_, err := godirwalk.Recurse(dirname, &godirwalk.RecurseOptions{
		Callback: func(osPathname string, de *godirwalk.Dirent, siblingSize, childSize int64) (int64, error) {
			if de.IsDir() {
				fmt.Printf("%d	%s\n", childSize, osPathname)
			}
			file, _ := os.Stat(osPathname)
			return siblingSize+childSize+file.Size(), nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			// Your program may want to log the error somehow.
			// fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)

			// For the purposes of this example, a simple SkipNode will suffice,
			// although in reality perhaps additional logic might be called for.
			return godirwalk.SkipNode
		},
		Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
	}, int64(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
