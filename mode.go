package godirwalk

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

// WalkFileModeFunc is the type of the function called for each file system node
// visited by WalkFileMode. The path argument contains the argument to
// WalkFileMode as a prefix; that is, if WalkFileMode is called with "dir",
// which is a directory containing the file "a", the walk function will be
// called with the argument "dir/a", using the correct os.PathSeparator for the
// Go Operating System architecture, GOOS. The mode argument is the os.FileMode
// for the named path, masked to the bits that identify the file system node
// type, i.e., os.ModeType.
//
// If an error is returned by the walk function, processing stops. The sole
// exception is when the function returns the special value filepath.SkipDir. If
// the function returns filepath.SkipDir when invoked on a directory,
// WalkFileMode skips the directory's contents entirely. If the function returns
// filepath.SkipDir when invoked on a non-directory file system node,
// WalkFileMode skips the remaining files in the containing directory.
type WalkFileModeFunc func(pathname string, mode os.FileMode) error

// WalkFileMode walks the file tree rooted at the specified directory, calling
// the specified callback function for each file system node in the tree,
// including root, symbolic links, and other node types. The nodes are walked in
// lexical order, which makes the output deterministic but means that for very
// large directories this function can be inefficient.
//
// This function is often much faster than filepath.Walk because it does not
// invoke os.Stat for every node it encounters, but rather gets the file system
// node type when it reads the parent directory.
func WalkFileMode(pathname string, walkFn WalkFileModeFunc) error {
	pathname = filepath.Clean(pathname)

	// Ensure specified pathname is a directory.
	fi, err := os.Stat(pathname)
	if err != nil {
		return errors.Wrap(err, "cannot Stat")
	}

	err = walker(pathname, fi.Mode()&os.ModeType, false, walkFn)
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

// WalkFileModeFollowSymlinks walks the file tree rooted at the specified
// directory, calling the specified callback function for each file system node
// in the tree, including root, symbolic links, and other node types. The nodes
// are walked in lexical order, which makes the output deterministic but means
// that for very large directories this function can be inefficient.
//
// This function is often much faster than filepath.Walk because it does not
// invoke os.Stat every node it encounters, but rather gets the file system node
// type when it reads the parent directory.
//
// This function also follows symbolic links that point to directories, and
// therefore ought to be used with caution, as calling it may cause an infinite
// loop in cases where the file system includes a logical loop of symbolic
// links.
func WalkFileModeFollowSymlinks(pathname string, walkFn WalkFileModeFunc) error {
	pathname = filepath.Clean(pathname)

	// Ensure specified pathname is a directory.
	fi, err := os.Stat(pathname)
	if err != nil {
		return errors.Wrap(err, "cannot read node")
	}

	err = walker(pathname, fi.Mode()&os.ModeType, true, walkFn)
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

func walker(osPathname string, modeType os.FileMode, followSymlinks bool, walkFn WalkFileModeFunc) error {
	err := walkFn(osPathname, modeType)
	if err != nil {
		if err != filepath.SkipDir {
			return errors.Wrap(err, "WalkFileModeFunc") // wrap error returned by walkFn
		}
		return err
	}

	if followSymlinks && modeType&os.ModeSymlink != 0 {
		// Resolve symbolic link referent to determine whether referent is
		// directory or not.
		fi, err := os.Stat(osPathname)
		if err != nil {
			return errors.Wrap(err, "cannot Stat")
		}
		modeType = fi.Mode() & os.ModeType
	}

	if modeType&os.ModeDir == 0 {
		return nil
	}

	// If get here, then specified pathname refers to a directory.
	deChildren, err := ReadDirents(osPathname, 0)
	if err != nil {
		return errors.Wrap(err, "cannot ReadDirents")
	}
	sort.Sort(deChildren)

	for _, deChild := range deChildren {
		osChildname := filepath.Join(osPathname, deChild.Name)
		err = walker(osChildname, deChild.ModeType, followSymlinks, walkFn)
		if err != nil {
			if err != filepath.SkipDir {
				return err
			}
			// If skipdir on a directory, stop processing that directory, but
			// continue to siblings. If skipdir on a non-directory, stop
			// processing siblings.
			if deChild.ModeType&os.ModeSymlink != 0 {
				// Resolve symbolic link referent to determine whether node
				// is directory or not.
				fi, err := os.Stat(osChildname)
				if err != nil {
					return errors.Wrap(err, "cannot Stat")
				}
				deChild.ModeType = fi.Mode() & os.ModeType
			}
			if deChild.ModeType&os.ModeDir == 0 {
				// If not directory, return immediately, thus skipping remainder
				// of siblings.
				return nil
			}
		}
	}
	return nil
}
