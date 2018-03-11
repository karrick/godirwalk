// +build darwin dragonfly linux nacl netbsd openbsd solaris

package godirwalk

import "syscall"

func direntIno(de *syscall.Dirent) uint64 {
	return de.Ino
}
