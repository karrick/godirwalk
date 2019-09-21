// +build aix js nacl solaris

package godirwalk

import (
	"os"
	"path/filepath"
	"syscall"
)

// modeType converts a syscall defined constant, which is in purview of OS, to a
// constant defined by Go, assumed by this project to be stable.
//
// Because some operating system syscall.Dirent structures do not include a Type
// field, fall back on Stat of the file system.
func modeType(_ *syscall.Dirent, osDirname, osBasename string) (os.FileMode, error) {
	return modeTypeLStat(filepath.Join(osDirname, osBasename))
}
