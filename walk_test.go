package godirwalk_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/karrick/godirwalk"
)

func helperFilepathWalk(tb testing.TB, osDirname string) []string {
	var entries []string
	err := filepath.Walk(osDirname, func(osPathname string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "skip" {
			return filepath.SkipDir
		}
		// filepath.Walk invokes callback function with a slashed version of the
		// pathname, while godirwalk invokes callback function with the
		// os-specific pathname separator.
		entries = append(entries, filepath.ToSlash(osPathname))
		return nil
	})
	if err != nil {
		tb.Fatal(err)
	}
	return entries
}

func helperGodirwalkWalk(tb testing.TB, osDirname string) []string {
	var entries []string
	err := godirwalk.Walk(osDirname, func(osPathname string, dirent *godirwalk.Dirent) error {
		if dirent.Name() == "skip" {
			return filepath.SkipDir
		}
		// filepath.Walk invokes callback function with a slashed version of the
		// pathname, while godirwalk invokes callback function with the
		// os-specific pathname separator.
		entries = append(entries, filepath.ToSlash(osPathname))
		return nil
	})
	if err != nil {
		tb.Fatal(err)
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

func BenchmarkFilepathWalk(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark using user's Go source directory")
	}

	prefix := filepath.Join(os.Getenv("GOPATH"), "src")

	for i := 0; i < b.N; i++ {
		_ = helperFilepathWalk(b, prefix)
	}
}

func BenchmarkGoDirWalk(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark using user's Go source directory")
	}

	prefix := filepath.Join(os.Getenv("GOPATH"), "src")

	for i := 0; i < b.N; i++ {
		_ = helperGodirwalkWalk(b, prefix)
	}
}
