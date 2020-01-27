package godirwalk

import "os"

func isDirectoryOrSymlinkToDirectory(de *Dirent, osPathname string) (bool, error) {
	if de.IsDir() {
		return true, nil
	}
	if !de.IsSymlink() {
		return false, nil
	}
	info, err := os.Stat(osPathname) // get information for referent
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
