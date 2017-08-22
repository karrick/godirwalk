package godirwalk

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

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
type WalkFileModeFunc func(osPathname string, mode os.FileMode) error

// WalkFileMode walks the file tree rooted at the specified directory, calling
// the specified callback function for each file system node in the tree,
// including root, symbolic links, and other node types. The nodes are walked in
// lexical order, which makes the output deterministic but means that for very
// large directories this function can be inefficient.
//
// This function is often much faster than filepath.Walk because it does not
// need to os.Stat every node it encounters, but rather gets the file system
// node type when it reads the parent directory.
func WalkFileMode(osDirname string, walkFn WalkFileModeFunc) error {
	return walk(osDirname, walkFn, false)
}

// WalkFileModeFollowSymlinks walks the file tree rooted at the specified
// directory, calling the specified callback function for each file system node
// in the tree, including root, symbolic links, and other node types. The nodes
// are walked in lexical order, which makes the output deterministic but means
// that for very large directories this function can be inefficient.
//
// This function is often much faster than filepath.Walk because it does not
// need to os.Stat every node it encounters, but rather gets the file system
// node type when it reads the parent directory.
//
// This function also follows symbolic links that point to directories, and
// ought to be used with caution, as calling it may cause an infinite loop in
// cases where the file system includes a logical loop of symbolic links.
func WalkFileModeFollowSymlinks(osDirname string, walkFn WalkFileModeFunc) error {
	return walk(osDirname, walkFn, true)
}

func walk(osDirname string, walkFn WalkFileModeFunc, followSymlinks bool) error {
	osDirname = filepath.Clean(osDirname)

	// Ensure parameter is a directory
	fi, err := os.Stat(osDirname)
	if err != nil {
		return errors.Wrap(err, "cannot read node")
	}
	if !fi.IsDir() {
		return errors.Errorf("cannot walk non directory: %q", osDirname)
	}

	// Initialize a work queue with the empty string, which signifies the
	// starting directory itself.
	de := &Dirent{Name: "", ModeType: os.ModeDir}
	queue := []*Dirent{de}

	scratchBuffer := make([]byte, 16*1024)

	// As we enumerate over the queue and encounter a directory, its children
	// will be added to the work queue.
	for len(queue) > 0 {
		// Unshift a pathname from the queue (breadth-first traversal of
		// hierarchy)
		de = queue[0]
		queue[0] = nil
		queue = queue[1:]

		osPathname := filepath.Join(osDirname, de.Name)

		if err = walkFn(osPathname, de.ModeType); err != nil {
			if err == filepath.SkipDir {
				if de.ModeType == os.ModeSymlink {
					// Resolve symbolic link referent to determine whether node
					// is directory or not.
					fi, err = os.Stat(osPathname)
					if err != nil {
						return errors.Wrap(err, "cannot stat")
					}
					if fi.IsDir() {
						de.ModeType = os.ModeDir
					}
				}
				// If current node is directory, then skip it; otherwise, skip
				// all nodes in the same parent directory.
				if de.ModeType != os.ModeDir {
					// Consume nodes from queue while they have the same parent
					// as the current node.
					osParent := filepath.Dir(osPathname) + osPathSeparator
					for len(queue) > 0 && strings.HasPrefix(queue[0].Name, osParent) {
						queue[0], queue = nil, queue[1:] // drop sibling entry from queue
					}
				}

				continue
			}
			return errors.Wrap(err, "DirWalkFileModeFunc") // wrap error returned by walkFn
		}

		if followSymlinks && de.ModeType == os.ModeSymlink {
			// Resolve symbolic link referent to determine whether node
			// is directory or not.
			fi, err = os.Stat(osPathname)
			if err != nil {
				return errors.Wrap(err, "cannot stat")
			}
			if fi.IsDir() {
				de.ModeType = os.ModeDir
			}
		}

		if de.ModeType == os.ModeDir {
			deChildren, err := GetDirectoryEntriesBuffer(osPathname, scratchBuffer)
			if err != nil {
				return errors.Wrap(err, "cannot get list of directory children")
			}
			sort.Sort(deChildren)
			for _, deChild := range deChildren {
				deChild.Name = filepath.Join(de.Name, deChild.Name)
				queue = append(queue, deChild)
			}
		}
	}
	de = nil
	return nil
}
