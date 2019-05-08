package godirwalk

// On Windows, a symbolic link to a directory already has mode type bits set for
// both directory and symbolic link.  This function need not do anything.
func isSymlinkToDirectory(_ string) (bool, error) { return false, nil }
