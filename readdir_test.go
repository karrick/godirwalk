package godirwalk

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestReadDirents(t *testing.T) {
	t.Skip("FIXME")

	actual, err := ReadDirents(testRoot, nil)
	ensureError(t, err)

	expected := Dirents{
		&Dirent{
			name:     "dir1",
			modeType: os.ModeDir,
		},
		&Dirent{
			name:     "dir2",
			modeType: os.ModeDir,
		},
		&Dirent{
			name:     "dir3",
			modeType: os.ModeDir,
		},
		&Dirent{
			name:     "dir4",
			modeType: os.ModeDir,
		},
		&Dirent{
			name:     "dir5",
			modeType: os.ModeDir,
		},
		&Dirent{
			name:     "dir6",
			modeType: os.ModeDir,
		},
		&Dirent{
			name:     "file3",
			modeType: 0,
		},
		&Dirent{
			name:     "symlinks",
			modeType: os.ModeDir,
		},
	}

	if got, want := len(actual), len(expected); got != want {
		t.Fatalf("(GOT) %v; (WNT) %v", got, want)
	}

	sort.Sort(actual)
	sort.Sort(expected)

	for i := 0; i < len(actual); i++ {
		if got, want := actual[i].name, expected[i].name; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
		if got, want := actual[i].modeType, expected[i].modeType; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
	}
}

func TestReadDirentsSymlinks(t *testing.T) {
	t.Skip("FIXME")
	osDirname := filepath.Join(testRoot, "symlinks")

	// Because some platforms set multiple mode type bits, when we create the
	// expected slice, we need to ensure the mode types are set appropriately.
	var expected Dirents
	for _, pathname := range []string{"aaa.txt", "symlinkToAbsDirectory", "symlinkToDirectory", "symlinkToFile", "symlinkToNothing", "zzz"} {
		info, err := os.Lstat(filepath.Join(osDirname, pathname))
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, &Dirent{name: pathname, modeType: info.Mode() & os.ModeType})
	}

	actual, err := ReadDirents(osDirname, nil)
	ensureError(t, err)

	if got, want := len(actual), len(expected); got != want {
		t.Fatalf("(GOT) %v; (WNT) %v", got, want)
	}

	sort.Sort(actual)
	sort.Sort(expected)

	for i := 0; i < len(actual); i++ {
		if got, want := actual[i].name, expected[i].name; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
		if got, want := actual[i].modeType, expected[i].modeType; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
	}
}

func TestReadDirnames(t *testing.T) {
	t.Skip("FIXME")
	actual, err := ReadDirnames(testRoot, nil)
	ensureError(t, err)
	expected := []string{"dir1", "dir2", "dir3", "symlinks", "dir5", "dir6", "dir4", "file3"}
	ensureStringSlicesMatch(t, actual, expected)
}
