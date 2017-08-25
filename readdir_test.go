package godirwalk

import (
	"os"
	"sort"
	"testing"
)

func TestReaddirents(t *testing.T) {
	entries, err := ReadDirents("testdata", 0)
	if err != nil {
		t.Fatal(err)
	}

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
			name:     "file3",
			modeType: 0,
		},
		&Dirent{
			name:     "symlinks",
			modeType: os.ModeDir,
		},
	}

	if got, want := len(entries), len(expected); got != want {
		t.Fatalf("(GOT) %v; (WNT) %v", got, want)
	}

	sort.Sort(entries)
	sort.Sort(expected)

	for i := 0; i < len(entries); i++ {
		if got, want := entries[i].name, expected[i].name; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
		if got, want := entries[i].modeType, expected[i].modeType; got != want {
			t.Errorf("(GOT) %v; (WNT) %v", got, want)
		}
	}
}

func TestReaddirnames(t *testing.T) {
	entries, err := ReadDirnames("testdata", 0)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"dir1", "dir2", "file3", "symlinks"}

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
