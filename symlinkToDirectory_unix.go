// +build !windows

package godirwalk

import "os"

// On UNIX, a symbolic link to a directory only has mode type bit for a symbolic
// link.  Must Stat entry to know whether or not its referent is a directory.
func isSymlinkToDirectory(osPathname string) (bool, error) {
	info, err := os.Stat(osPathname)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
