package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := filepath.Walk(filepath.Clean(os.Args[1]), callback); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func callback(osPathname string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	fmt.Printf("%s %s\n", info.Mode(), osPathname)
	return nil
}
