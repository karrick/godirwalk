package godirwalk

import "os"

func isDirectoryOrSymlinkToDirectory(de *Dirent, osPathname string) (bool, error) {
	if de.IsDir() {
		return true, nil
	}
	if !de.IsSymlink() {
		return false, nil
	}
	// Does this symlink point to a directory?
	info, err := os.Stat(osPathname)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
