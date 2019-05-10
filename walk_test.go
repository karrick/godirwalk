package godirwalk

import (
	"os"
	"path/filepath"
	"testing"
)

func filepathWalk(tb testing.TB, osDirname string) []string {
	tb.Helper()
	var entries []string
	err := filepath.Walk(osDirname, func(osPathname string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "skip" {
			return filepath.SkipDir
		}
		entries = append(entries, filepath.FromSlash(osPathname))
		return nil
	})
	if err != nil {
		tb.Fatal(err)
	}
	return entries
}

func godirwalkWalk(tb testing.TB, osDirname string) []string {
	tb.Helper()
	var entries []string
	err := Walk(osDirname, &Options{
		ScratchBuffer: testScratchBuffer,
		Callback: func(osPathname string, dirent *Dirent) error {
			if dirent.Name() == "skip" {
				return filepath.SkipDir
			}
			entries = append(entries, filepath.FromSlash(osPathname))
			return nil
		},
	})
	if err != nil {
		tb.Fatal(err)
	}
	return entries
}

// Ensure the results from calling this library's Walk function exactly match
// those returned by filepath.Walk
func ensureSameAsStandardLibrary(tb testing.TB, osDirname string) {
	tb.Helper()
	osDirname = filepath.Join(testRoot, osDirname)
	actual := godirwalkWalk(tb, osDirname)
	expected := filepathWalk(tb, osDirname)
	ensureStringSlicesMatch(tb, actual, expected)
}

// Test the entire test root hierarchy with all of its artifacts.  This library
// advertises itself as visiting the same file system entries as the standard
// library, and responding to discovered errors the same way, including
// responding to filepath.SkipDir exactly like the standard library does.  This
// test ensures that behavior is correct by enumerating the contents of the test
// root directory.
func TestWalkCompatibleWithFilepathWalk(t *testing.T) {
	t.Run("test root", func(t *testing.T) {
		ensureSameAsStandardLibrary(t, "dir1")
	})
}

// While filepath.Walk will deliver the no access error to the regular callback,
// godirwalk should deliver it first to the ErrorCallback handler, then take
// action based on the return value of that callback function.
func TestErrorCallback(t *testing.T) {
	t.Run("halt", func(t *testing.T) {
		var callbackVisited, errorCallbackVisited bool

		err := Walk(filepath.Join(testRoot, "noaccess/dir"), &Options{
			ScratchBuffer: testScratchBuffer,
			Callback: func(osPathname string, dirent *Dirent) error {
				switch dirent.Name() {
				case "never":
					t.Errorf("Callback VISITED: %v", osPathname)
				case "trap":
					callbackVisited = true
				}
				return nil
			},
			ErrorCallback: func(osPathname string, err error) ErrorAction {
				switch filepath.Base(osPathname) {
				case "trap":
					errorCallbackVisited = true
					return Halt // Direct Walk to propagate error to caller
				}
				t.Fatalf("unexpected error callback for %s: %s", osPathname, err)
				return SkipNode
			},
		})

		ensureError(t, err, "trap") // Ensure caller receives propagated access error
		if got, want := callbackVisited, true; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := errorCallbackVisited, true; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
	})

	t.Run("skipnode", func(t *testing.T) {
		var callbackVisited, errorCallbackVisited bool

		err := Walk(filepath.Join(testRoot, "noaccess/dir"), &Options{
			ScratchBuffer: testScratchBuffer,
			Callback: func(osPathname string, dirent *Dirent) error {
				switch dirent.Name() {
				case "never":
					t.Errorf("Callback VISITED: %v", osPathname)
				case "trap":
					callbackVisited = true
				}
				return nil
			},
			ErrorCallback: func(osPathname string, err error) ErrorAction {
				switch filepath.Base(osPathname) {
				case "trap":
					errorCallbackVisited = true
					return SkipNode // Direct Walk to ignore this error
				}
				t.Fatalf("unexpected error callback for %s: %s", osPathname, err)
				return Halt
			},
		})

		ensureError(t, err) // Ensure caller receives no access error
		if got, want := callbackVisited, true; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := errorCallbackVisited, true; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
	})
}

// Test cases for encountering the filepath.SkipDir error at different
// relative positions from the invocation argument.
func TestWalkSkipDir(t *testing.T) {
	t.Run("SkipFileAtRoot", func(t *testing.T) {
		ensureSameAsStandardLibrary(t, "dir1/dir1a")
	})

	t.Run("SkipFileUnderRoot", func(t *testing.T) {
		ensureSameAsStandardLibrary(t, "dir1")
	})

	t.Run("SkipDirAtRoot", func(t *testing.T) {
		ensureSameAsStandardLibrary(t, "dir2/skip")
	})

	t.Run("SkipDirUnderRoot", func(t *testing.T) {
		ensureSameAsStandardLibrary(t, "dir2")
	})

	t.Run("SkipDirOnSymlink", func(t *testing.T) {
		var actual []string
		err := Walk(filepath.Join(testRoot, "dir3"), &Options{
			ScratchBuffer: testScratchBuffer,
			Callback: func(osPathname string, dirent *Dirent) error {
				if dirent.Name() == "skip" {
					return filepath.SkipDir
				}
				actual = append(actual, filepath.FromSlash(osPathname))
				return nil
			},
			FollowSymbolicLinks: true, // make sure it normally would follow the links
		})
		if err != nil {
			t.Fatal(err)
		}

		expected := []string{
			filepath.Join(testRoot, "dir3"),
			filepath.Join(testRoot, "dir3/aaa.txt"),
			filepath.Join(testRoot, "dir3/zzz"),
			filepath.Join(testRoot, "dir3/zzz/aaa.txt"),
		}

		ensureStringSlicesMatch(t, actual, expected)
	})
}

func TestWalkFollowSymbolicLinks(t *testing.T) {
	var actual []string
	err := Walk(filepath.Join(testRoot, "symlinks"), &Options{
		ScratchBuffer: testScratchBuffer,
		Callback: func(osPathname string, _ *Dirent) error {
			actual = append(actual, filepath.FromSlash(osPathname))
			return nil
		},
		FollowSymbolicLinks: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		filepath.Join(testRoot, "symlinks"),
		filepath.Join(testRoot, "symlinks/aaa.txt"),
		filepath.Join(testRoot, "symlinks/symlinkToAbsDirectory"),
		filepath.Join(testRoot, "symlinks/symlinkToAbsDirectory/aaa.txt"),
		filepath.Join(testRoot, "symlinks/symlinkToDirectory"),
		filepath.Join(testRoot, "symlinks/symlinkToDirectory/aaa.txt"),
		filepath.Join(testRoot, "symlinks/symlinkToFile"),
		filepath.Join(testRoot, "symlinks/zzz"),
		filepath.Join(testRoot, "symlinks/zzz/aaa.txt"),
	}

	ensureStringSlicesMatch(t, actual, expected)
}

func TestWalkSymbolicRelativeLinkChain(t *testing.T) {
	var actual []string
	err := Walk(filepath.Join(testRoot, "dir4"), &Options{
		ScratchBuffer: testScratchBuffer,
		Callback: func(osPathname string, _ *Dirent) error {
			actual = append(actual, filepath.FromSlash(osPathname))
			return nil
		},
		FollowSymbolicLinks: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		filepath.Join(testRoot, "dir4"),
		filepath.Join(testRoot, "dir4", "a"),
		filepath.Join(testRoot, "dir4", "a", "x"),
		filepath.Join(testRoot, "dir4", "a", "x", "y"),
		filepath.Join(testRoot, "dir4", "b"),
		filepath.Join(testRoot, "dir4", "b", "y"),
		filepath.Join(testRoot, "dir4", "z"),
	}

	ensureStringSlicesMatch(t, actual, expected)
}

func TestPostChildrenCallback(t *testing.T) {
	osDirname := filepath.Join(testRoot, "dir5")

	var actual []string

	err := Walk(osDirname, &Options{
		ScratchBuffer: testScratchBuffer,
		Callback: func(_ string, _ *Dirent) error {
			return nil
		},
		PostChildrenCallback: func(osPathname string, _ *Dirent) error {
			actual = append(actual, osPathname)
			return nil
		},
	})
	if err != nil {
		t.Errorf("(GOT): %v; (WNT): %v", err, nil)
	}

	expected := []string{
		filepath.Join(testRoot, "dir5/a2/a2a"),
		filepath.Join(testRoot, "dir5/a2"),
		filepath.Join(testRoot, "dir5"),
	}

	ensureStringSlicesMatch(t, actual, expected)
}

var goPrefix = filepath.Join(os.Getenv("GOPATH"), "src")

func BenchmarkFilepathWalk(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark using user's Go source directory")
	}

	for i := 0; i < b.N; i++ {
		_ = filepathWalk(b, goPrefix)
	}
}

func BenchmarkGodirwalk(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark using user's Go source directory")
	}

	for i := 0; i < b.N; i++ {
		_ = godirwalkWalk(b, goPrefix)
	}
}

const flameIterations = 10

func BenchmarkFlameGraphFilepathWalk(b *testing.B) {
	for i := 0; i < flameIterations; i++ {
		_ = filepathWalk(b, goPrefix)
	}
}

func BenchmarkFlameGraphGodirwalk(b *testing.B) {
	for i := 0; i < flameIterations; i++ {
		_ = godirwalkWalk(b, goPrefix)
	}
}
