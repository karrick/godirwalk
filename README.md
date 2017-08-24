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

There are two major differences and one minor difference to the
operation of `filepath.Walk` and the directory traversal algorithm in
this library.

First, `filepath.Walk` invokes the callback function with a slashed
version of the pathname regardless of the os-specific path separator,
while godirwalk invokes callback function with the os-specific
pathname separator.

Second, while `filepath.Walk` invokes callback function with the
`os.FileInfo` for every node, this library invokes the callback
function with the `os.FileMode` set to the type of file system node it
is, namely, by masking the file system mode with `os.ModeType`. It
does this because this eliminates the need to invoke `os.Stat` on
every file system node. On the occassion that the callback function
needs the full stat information, it can call `os.Stat` when required.

Third, since this library does not invoke `os.Stat` on every node,
there is no possible error event for the callback function to filter
on. The third argument in the callback function signature for the stat
error is no longer necessary.

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
	err := godirwalk.Walk(dirname, callback)
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
