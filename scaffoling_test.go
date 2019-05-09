package godirwalk

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const testScratchBufferSize = 16 * 1024

var testScratchBuffer []byte
var testRoot string

func init() {
	testScratchBuffer = make([]byte, testScratchBufferSize)
}

func TestMain(m *testing.M) {
	flag.Parse()

	var code int // program exit code

	// All tests use the same directory test scaffolding.  Create the directory
	// hierarchy, run the tests, then remove the root directory of the test
	// scaffolding.

	defer func() {
		if err := teardown(); err != nil {
			fmt.Fprintf(os.Stderr, "godirwalk teardown: %s\n", err)
			code = 1
		}
		os.Exit(code)
	}()

	// When cannot complete setup, dump the directory so we see what we have,
	// then bail.
	if err := setup(); err != nil {
		fmt.Fprintf(os.Stderr, "godirwalk setup: %s\n", err)
		dumpDirectory()
		code = 1
		return
	}

	code = m.Run()

	// When any test was a failure, then use standard library to walk test
	// scaffolding directory and print its contents.
	if code != 0 {
		dumpDirectory()
	}
}

func setup() error {
	var err error

	testRoot, err = ioutil.TempDir(os.TempDir(), "godirwalk-")
	if err != nil {
		return err
	}

	// Create files, creating parent directories along the way.
	files := []string{
		"dir1/dir1a/file1a1",
		"dir1/dir1a/skip",
		"dir1/dir1a/z1a2",
		"dir1/file1b",
		"dir2/file2a",
		"dir2/skip/file2b1",
		"dir2/z2c/file2c1",
		"dir3/aaa.txt",
		"dir3/zzz/aaa.txt",
		"dir4/aaa.txt",
		"dir4/zzz/aaa.txt",
		"dir5/a1.txt",
		"dir5/a2/a2a/a2a1.txt",
		"dir5/a2/a2b.txt",
		"dir6/bravo.txt",
		"dir6/code/123.txt",
		"dir7/z",
		"file3",
	}

	for _, newname := range files {
		newname = filepath.Join(testRoot, filepath.FromSlash(newname))
		if err := os.MkdirAll(filepath.Dir(newname), os.ModePerm); err != nil {
			return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
		}
		if err = ioutil.WriteFile(newname, []byte("some test data\n"), os.ModePerm); err != nil {
			return fmt.Errorf("cannot create file for test scaffolding: %s", err)
		}
	}

	// Create a directory for which the testing user has no access.
	if err := os.MkdirAll(filepath.Join(testRoot, filepath.FromSlash("dir6/noaccess/dir")), os.FileMode(0)); err != nil {
		return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
	}

	// Create an symbolic link to an absolute pathname.
	oldname, err := filepath.Abs(filepath.Join(testRoot, "dir4/zzz"))
	if err != nil {
		return fmt.Errorf("cannot create absolute pathname for test scaffolding: %s", err)
	}
	newname := filepath.Join(testRoot, "dir4/symlinkToAbsDirectory")
	if err := os.Symlink(oldname, newname); err != nil {
		return fmt.Errorf("cannot create symlink to absolute directory for test scaffolding: %s", err)
	}

	// Create a handful of symbolic links, creating parent directories along the
	// way.
	symlinks := []struct {
		newname, oldname string
	}{
		// {"dir3/skip", "zzz"},
		{"dir4/skip", "zzz"},
		{"dir4/symlinkToDirectory", "zzz"},
		{"dir4/symlinkToFile", "aaa.txt"},
		// {"dir4/symlinkToInfinity", "../dir4"},
		{"dir4/symlinkToNothing", "non-existing-file"},
		{"dir4/b/y", "../z"},
		{"dir4/a/x", "../b"},
	}

	for _, entry := range symlinks {
		newname := filepath.Join(testRoot, filepath.FromSlash(entry.newname))
		if err := os.MkdirAll(filepath.Dir(newname), os.ModePerm); err != nil {
			return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
		}
		oldname := filepath.FromSlash(entry.oldname)
		if err := os.Symlink(oldname, newname); err != nil {
			return fmt.Errorf("cannot create symbolic link for test scaffolding: %s", err)
		}
	}

	// Create a few empty directory entries.
	extraDirs := []string{
		"dir6/abc",
		"dir6/def",
	}

	for _, newname := range extraDirs {
		newname = filepath.Join(testRoot, filepath.FromSlash(newname))
		if err := os.MkdirAll(newname, os.ModePerm); err != nil {
			return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
		}
	}

	return nil
}

func teardown() error {
	if testRoot == "" {
		return nil // if we do not even have a test root directory then exit
	}
	// Change permissions back to something we will later be permitted to delete.
	if err := os.Chmod(filepath.Join(testRoot, filepath.FromSlash("dir6/noaccess")), os.ModePerm); err != nil {
		return fmt.Errorf("cannot change permission to delete dir6/noaccess for test scaffolding: %s", err)
	}
	if err := os.RemoveAll(testRoot); err != nil {
		return err
	}
	return nil
}

func dumpDirectory() {
	trim := len(testRoot) // trim rootDir from prefix of strings
	err := filepath.Walk(testRoot, func(osPathname string, info os.FileInfo, err error) error {
		if err != nil {
			// we have no info, so get it
			info, err2 := os.Lstat(osPathname)
			if err2 != nil {
				fmt.Fprintf(os.Stderr, "?--------- %s: %s\n", osPathname[trim:], err2)
			} else {
				fmt.Fprintf(os.Stderr, "%s %s: %s\n", info.Mode(), osPathname[trim:], err)
			}
			return nil
		}

		var suffix string

		if info.Mode()&os.ModeSymlink != 0 {
			referent, err := os.Readlink(osPathname)
			if err != nil {
				suffix = fmt.Sprintf(": cannot read symlink: %s", err)
				err = nil
			} else {
				suffix = fmt.Sprintf(" -> %s", referent)
			}
		}
		fmt.Fprintf(os.Stderr, "%s %s%s\n", info.Mode(), osPathname[trim:], suffix)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot walk test directory: %s\n", err)
	}
}
