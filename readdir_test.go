package godirwalk_test

import (
	"os"
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
