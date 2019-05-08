package godirwalk

func isDirectoryOrSymlinkToDirectory(de *Dirent, osPathname string) (bool, error) {
	// On Windows, a symbolic link to a directory has both the symbolic link and
	// the directory mode bits set.
	if de.IsDir() {
		return true, nil
	}
	if !de.IsSymlink() {
		return false, nil
	}
	// On UNIX, a symbolic link to a directory needs to be followed to determine
	// its referent.
	return isSymlinkToDirectory(osPathname)
}
