package godirwalk_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/karrick/godirwalk"
)

func TestReaddirnames(t *testing.T) {
	entries, err := godirwalk.ReadDirnames("testdata", 0)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"dir1", "dir2", "file3"}

	if got, want := len(entries), len(expected); got != want {
		t.Fatalf("(GOT) %v; (WNT) %v", got, want)
	}

	sort.Strings(entries)
	sort.Strings(expected)

	for i := 0; i < len(entries); i++ {
		if got, want := entries[i], expected[i]; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
	}
}

func TestReaddirents(t *testing.T) {
	entries, err := godirwalk.ReadDirents("testdata", 0)
	if err != nil {
		t.Fatal(err)
	}

	expected := godirwalk.Dirents{
		&godirwalk.Dirent{
			Name:     "dir1",
			ModeType: os.ModeDir,
		},
		&godirwalk.Dirent{
			Name:     "dir2",
			ModeType: os.ModeDir,
		},
		&godirwalk.Dirent{
			Name:     "file3",
			ModeType: 0,
		},
	}

	if got, want := len(entries), len(expected); got != want {
		t.Fatalf("(GOT) %v; (WNT) %v", got, want)
	}

	sort.Sort(entries)
	sort.Sort(expected)

	for i := 0; i < len(entries); i++ {
		if got, want := entries[i].Name, expected[i].Name; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
		if got, want := entries[i].ModeType, expected[i].ModeType; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
	}
}

func helperFilepathWalk(t *testing.T, osDirname string) []string {
	var entries []string
	err := filepath.Walk(osDirname, func(osPathname string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(osPathname) == "skip" {
			return filepath.SkipDir
		}
		// filepath.Walk invokes callback function with a slashed version of the
		// pathname, while godirwalk invokes callback function with the
		// os-specific pathname separator.
		entries = append(entries, filepath.ToSlash(osPathname))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func helperGodirwalkWalk(t *testing.T, osDirname string) []string {
	var entries []string
	err := godirwalk.Walk(osDirname, func(osPathname string, _ os.FileMode) error {
		if filepath.Base(osPathname) == "skip" {
			return filepath.SkipDir
		}
		// filepath.Walk invokes callback function with a slashed version of the
		// pathname, while godirwalk invokes callback function with the
		// os-specific pathname separator.
		entries = append(entries, filepath.ToSlash(osPathname))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func TestWalkSkipDir(t *testing.T) {
	// Ensure the results from calling filepath.Walk exactly match the results
	// for calling this library's walk function.
	test := func(t *testing.T, osDirname string) {
		expected := helperFilepathWalk(t, osDirname)
		actual := helperGodirwalkWalk(t, osDirname)

		if got, want := len(actual), len(expected); got != want {
			t.Fatalf("\n(GOT)\n\t%#v\n(WNT)\n\t%#v", actual, expected)
		}

		for i := 0; i < len(actual); i++ {
			if got, want := actual[i], expected[i]; got != want {
				t.Errorf("(GOT) %v; (WNT) %v", got, want)
			}
		}
	}

	// Test cases for encountering the filepath.SkipDir error at different times
	// from the call.

	t.Run("SkipFileAtRoot", func(t *testing.T) {
		test(t, "testdata/dir1/dir1a")
	})

	t.Run("SkipFileUnderRoot", func(t *testing.T) {
		test(t, "testdata/dir1")
	})

	t.Run("SkipDirAtRoot", func(t *testing.T) {
		test(t, "testdata/dir2/skip")
	})

	t.Run("SkipDirUnderRoot", func(t *testing.T) {
		test(t, "testdata/dir2")
	})
}
