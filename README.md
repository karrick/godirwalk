# godirwalk

`godirwalk` is a library for walking the file system.

## Description

The Go standard library provides a flexible function for traversing a
file system directory tree. However it invokes `os.Stat` on every file
system node encountered in order to provide the file info data
structure, i.e. `os.FileInfo`, to the upstream client. For many uses
the `os.Stat` is needless and adversely impacts performance. Many
clients need to branch based on file system node type, and, that
information is already provided by the system call used to read a
directory's children nodes.

## Usage

Documentation is available via
[![GoDoc](https://godoc.org/github.com/karrick/godirwalk?status.svg)](https://godoc.org/github.com/karrick/godirwalk).

```Go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

func main() {
	dirname := "."
	if len(os.Args) > 1 {
		dirname = os.Args[1]
	}
	err := godirwalk.WalkFileMode(filepath.Clean(dirname), callback)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}

func callback(osPathname string, mode os.FileMode) error {
	fmt.Printf("%s %s\n", mode, osPathname)
	return nil
}
```
