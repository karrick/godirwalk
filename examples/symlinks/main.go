package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
)

func main() {
	dirname := "."
	if len(os.Args) > 1 {
		dirname = os.Args[1]
	}
	i := 0
	err := godirwalk.Walk(dirname, &godirwalk.Options{
		FollowSymbolicLinks: true,
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			i++
			fmt.Println(i, ":", osPathname)
			if de.IsSymlink() {
				target, err := os.Readlink(osPathname)
				if err != nil {
					return err
				}
				if target == "." || strings.HasPrefix(target, "..") || filepath.Base(filepath.Clean(osPathname)) == target {
					fmt.Println(i, ": Skipping target", target)
					return filepath.SkipDir
					//return nil
				}
			}
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			panic(fmt.Sprintf("ERROR: %s : %s\n", err.Error(), osPathname))
			return godirwalk.SkipNode
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
