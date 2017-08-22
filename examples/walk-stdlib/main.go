package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	err := filepath.Walk(filepath.Clean(os.Args[1]), func(osPathname string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", osPathname)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}
