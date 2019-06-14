/*
 * sizes
 *
 * Walks a file system hierarchy and prints sizes of file system objects,
 * recursively printing sizes of directories.
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

var progname = filepath.Base(os.Args[0])

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s dir1 [dir2 [dir3...]]\n", progname)
		os.Exit(2)
	}

	scratchBuffer := make([]byte, 64*1024) // allocate once and re-use each time

	for _, arg := range os.Args[1:] {
		if err := sizes(arg, scratchBuffer); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", progname, err)
		}
	}
}

func sizes(osDirname string, scratchBuffer []byte) error {
	var sizes []int64 // LIFO (push, pop semantics)

	return godirwalk.Walk(osDirname, &godirwalk.Options{
		ScratchBuffer: scratchBuffer,
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				sizes = append(sizes, 0)
				return nil
			}

			st, err := os.Stat(osPathname)
			if err != nil {
				return err
			}

			size := st.Size()
			sizes[len(sizes)-1] += size

			_, err = fmt.Printf("%s % 12d %s\n", st.Mode(), size, osPathname)
			return err
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			fmt.Fprintf(os.Stderr, "%s: %s\n", progname, err)
			return godirwalk.SkipNode
		},
		PostChildrenCallback: func(osPathname string, de *godirwalk.Dirent) error {
			i := len(sizes) - 1
			var size int64
			size, sizes = sizes[i], sizes[:i]

			st, err := os.Stat(osPathname)

			switch err {
			case nil:
				_, err = fmt.Printf("%s % 12d %s\n", st.Mode(), size, osPathname)
			default:
				// ignore the error and just show the mode type
				_, err = fmt.Printf("%s % 12d %s\n", de.ModeType(), size, osPathname)
			}
			return err
		},
	})
}
