package godirwalk

import (
	"os"
	"path/filepath"
	"testing"
)

func readdirentsUsingScan(osDirname string, scratchBuffer []byte) (Dirents, error) {
	var entries Dirents
	scanner, err := NewDirectoryScanner(osDirname, scratchBuffer)
	if err != nil {
		return nil, err
	}
	for scanner.Scan() {
		entry, err := scanner.Entry()
		_ = err // ignore error from entry
		entries = append(entries, entry.Dup())
	}
	if err = scanner.Close(); err != nil {
		return nil, err
	}
	return entries, nil
}

func readdirnamesUsingScan(osDirname string, scratchBuffer []byte) ([]string, error) {
	var entries []string
	scanner, err := NewDirectoryScanner(osDirname, scratchBuffer)
	if err != nil {
		return nil, err
	}
	for scanner.Scan() {
		entry, err := scanner.Entry()
		_ = err // ignore error from entry
		entries = append(entries, entry.Name())
	}
	if err = scanner.Close(); err != nil {
		return nil, err
	}
	return entries, nil
}

func TestScanDir(t *testing.T) {
	t.Run("dirents", func(t *testing.T) {
		actual, err := readdirentsUsingScan(filepath.Join(testRoot, "d0"), nil)

		ensureError(t, err)

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

	t.Run("dirnames", func(t *testing.T) {
		actual, err := readdirnamesUsingScan(filepath.Join(testRoot, "d0"), nil)
		ensureError(t, err)
		expected := []string{maxName, "d1", "f1", "skips", "symlinks"}
		ensureStringSlicesMatch(t, actual, expected)
	})
}
