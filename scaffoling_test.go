package godirwalk

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var rootDir string

func TestMain(m *testing.M) {
	flag.Parse()

	// All tests use the same directory test scaffolding.  Create the directory
	// hierarchy, run the tests, then remove the root directory of the test
	// scaffolding.
	if err := setup(); err != nil {
		fmt.Fprintf(os.Stderr, "setup: %s\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := teardown(); err != nil {
		fmt.Fprintf(os.Stderr, "teardown: %s\n", err)
		os.Exit(1)
	}

	os.Exit(code)
}

func setup() error {
	var err error

	rootDir, err = ioutil.TempDir(os.TempDir(), "godirwalk-")
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

	for _, pathname := range files {
		pathname = filepath.Join(rootDir, filepath.FromSlash(pathname))

		if err := os.MkdirAll(filepath.Dir(pathname), os.ModePerm); err != nil {
			return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
		}

		if err = ioutil.WriteFile(pathname, []byte("some test data\n"), os.ModePerm); err != nil {
			return fmt.Errorf("cannot create file for test scaffolding: %s", err)
		}
	}

	symlinks := []struct {
		newname, oldname string
	}{
		{"dir3/skip", "zzz"},
		{"dir4/symlinkToDirectory", "zzz"},
		{"dir4/symlinkToFile", "aaa.txt"},
		{"dir7/b/y", "../z"},
		{"dir7/a/x", "../b"},
		{"symlinks/dir-symlink", "../symlinks"}, // infinite loop of symlinks
		{"symlinks/file-symlink", "../file3"},
		{"symlinks/invalid-symlink", "/non/existing/file"},
	}

	for _, entry := range symlinks {
		newname := filepath.Join(rootDir, filepath.FromSlash(entry.newname))

		if err := os.MkdirAll(filepath.Dir(newname), os.ModePerm); err != nil {
			return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
		}

		oldname := filepath.FromSlash(entry.oldname)

		if err := os.Symlink(oldname, newname); err != nil {
			return fmt.Errorf("cannot create symbolic link for test scaffolding: %s", err)
		}
	}

	extraDirs := []string{
		"dir6/abc",
		"dir6/def",
	}

	for _, pathname := range extraDirs {
		pathname = filepath.Join(rootDir, filepath.FromSlash(pathname))

		if err := os.MkdirAll(pathname, os.ModePerm); err != nil {
			return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
		}
	}

	if err := os.MkdirAll(filepath.Join(rootDir, filepath.FromSlash("dir6/noaccess")), os.FileMode(0)); err != nil {
		return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
	}

	return nil
}

func teardown() error {
	if err := os.Chmod(filepath.Join(rootDir, filepath.FromSlash("dir6/noaccess")), os.ModePerm); err != nil {
		return fmt.Errorf("cannot change permission to delete dir6/noaccess for test scaffolding: %s", err)
	}
	if err := os.RemoveAll(rootDir); err != nil {
		return err
	}
	return nil
}
