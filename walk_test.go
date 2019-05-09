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
		ensureSameAsStandardLibrary(t, "")
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
		filepath.Join(testRoot, "dir4/aaa.txt"),
		filepath.Join(testRoot, "dir4/symlinkToAbsDirectory"),
		filepath.Join(testRoot, "dir4/symlinkToAbsDirectory/aaa.txt"),
		filepath.Join(testRoot, "dir4/symlinkToDirectory"),
		filepath.Join(testRoot, "dir4/symlinkToDirectory/aaa.txt"),
		filepath.Join(testRoot, "dir4/symlinkToFile"),
		filepath.Join(testRoot, "dir4/zzz"),
		filepath.Join(testRoot, "dir4/zzz/aaa.txt"),
	}

	ensureStringSlicesMatch(t, actual, expected)
}

func TestWalkSymbolicRelativeLinkChain(t *testing.T) {
	var actual []string
	err := Walk(filepath.Join(testRoot, "dir7"), &Options{
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
		filepath.Join(testRoot, "dir7"),
		filepath.Join(testRoot, "dir7", "a"),
		filepath.Join(testRoot, "dir7", "a", "x"),
		filepath.Join(testRoot, "dir7", "a", "x", "y"),
		filepath.Join(testRoot, "dir7", "b"),
		filepath.Join(testRoot, "dir7", "b", "y"),
		filepath.Join(testRoot, "dir7", "z"),
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
