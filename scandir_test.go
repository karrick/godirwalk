package godirwalk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScannerDirent(t *testing.T) {
	scanner, err := NewScanner(filepath.Join(testRoot, "d0/symlinks"))
	ensureError(t, err)

	var found bool

	for scanner.Scan() {
		if scanner.Name() != "toD1" {
			continue
		}
		found = true

		de, err := scanner.Dirent()
		ensureError(t, err)

		got, err := de.IsDirOrSymlinkToDir()
		ensureError(t, err)

		if want := true; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
	}

	ensureError(t, scanner.Err())

	if got, want := found, true; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestScanDir(t *testing.T) {
	t.Run("dirent", func(t *testing.T) {
		var actual []*Dirent

		scanner, err := NewScanner(filepath.Join(testRoot, "d0"))
		ensureError(t, err)

		for scanner.Scan() {
			dirent, err := scanner.Dirent()
			ensureError(t, err)
			actual = append(actual, dirent)
		}
		ensureError(t, scanner.Err())

		expected := Dirents{
			&Dirent{
				name:     maxName,
				modeType: os.FileMode(0),
			},
			&Dirent{
				name:     "d1",
				modeType: os.ModeDir,
			},
			&Dirent{
				name:     "f1",
				modeType: os.FileMode(0),
			},
			&Dirent{
				name:     "skips",
				modeType: os.ModeDir,
			},
			&Dirent{
				name:     "symlinks",
				modeType: os.ModeDir,
			},
		}

		ensureDirentsMatch(t, actual, expected)
	})

	t.Run("name", func(t *testing.T) {
		var actual []string

		scanner, err := NewScanner(filepath.Join(testRoot, "d0"))
		ensureError(t, err)

		for scanner.Scan() {
			actual = append(actual, scanner.Name())
		}
		ensureError(t, scanner.Err())

		expected := []string{maxName, "d1", "f1", "skips", "symlinks"}
		ensureStringSlicesMatch(t, actual, expected)
	})
}
