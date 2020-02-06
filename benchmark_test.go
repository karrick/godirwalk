package godirwalk

import (
	"path/filepath"
	"testing"
)

const benchRoot = "/mnt/ram_disk/src"

var scratch []byte
var largeDirectory string

func init() {
	scratch = make([]byte, MinimumScratchBufferSize)
	largeDirectory = filepath.Join(benchRoot, "linkedin/dashboards")
}

func Benchmark2ReadDirentsGodirwalk(b *testing.B) {
	var count int

	for i := 0; i < b.N; i++ {
		actual, err := ReadDirents(largeDirectory, scratch)
		if err != nil {
			b.Fatal(err)
		}
		count += len(actual)
	}

	_ = count
}

func Benchmark2ReadDirnamesGodirwalk(b *testing.B) {
	var count int

	for i := 0; i < b.N; i++ {
		actual, err := ReadDirnames(largeDirectory, scratch)
		if err != nil {
			b.Fatal(err)
		}
		count += len(actual)
	}

	_ = count
}

func Benchmark2GodirwalkSorted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var length int
		err := Walk(benchRoot, &Options{
			Callback: func(name string, _ *Dirent) error {
				if name == "skip" {
					return filepath.SkipDir
				}
				length += len(name)
				return nil
			},
			ScratchBuffer: scratch,
		})
		if err != nil {
			b.Errorf("GOT: %v; WANT: nil", err)
		}
		_ = length
	}
}

func Benchmark2GodirwalkUnsorted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var length int
		err := Walk(benchRoot, &Options{
			Callback: func(name string, _ *Dirent) error {
				if name == "skip" {
					return filepath.SkipDir
				}
				length += len(name)
				return nil
			},
			ScratchBuffer: scratch,
			Unsorted:      true,
		})
		if err != nil {
			b.Errorf("GOT: %v; WANT: nil", err)
		}
		_ = length
	}
}
