// +build dragonfly

package godirwalk

import "syscall"

func direntReclen(_ *syscall.Dirent, namelen uint64) uint64 {
	return (16 + namlen + 1 + 7) &^ 7
}
