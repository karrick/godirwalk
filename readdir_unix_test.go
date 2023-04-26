//go:build !windows
// +build !windows

package godirwalk

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestReadDirentsWithInodes(t *testing.T) {
	testroot := filepath.Join(scaffolingRoot, "d0")

	actual, err := ReadDirents(testroot, nil)
	ensureError(t, err)

	for n, entry := range actual {
		path := filepath.Join(testroot, entry.Name())
		stat, err := os.Stat(path)
		ensureError(t, err)

		statt, ok := stat.Sys().(*syscall.Stat_t)
		if !ok {
			t.Errorf("test %d: failed to get Stat_t", n+1)
			continue
		}
		if inode := entry.Inode(); inode != statt.Ino {
			t.Errorf("test %d: GOT: %d; WANT: %d", n+1, inode, statt.Ino)
		}
	}
}
