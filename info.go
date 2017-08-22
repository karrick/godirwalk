package godirwalk

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// WalkFileInfoFunc is the type of the function called for each file system node
// visited by WalkFileInfo. The path argument contains the argument to
// WalkFileInfo as a prefix; that is, if WalkFileInfo is called with "dir",
// which is a directory containing the file "a", the walk function will be
// called with the argument "dir/a", using the correct os.PathSeparator for the
// Go Operating System architecture, GOOS. The info argument is the os.FileInfo
// for the named path.
//
// If there was a problem walking to the file or directory named by path, the
// incoming error will describe the problem and the function can decide how to
// handle that error (and WalkFileInfo will not descend into that directory). If
// an error is returned, processing stops. The sole exception is when the
// function returns the special value filepath.SkipDir. If the function returns
// filepath.SkipDir when invoked on a directory, WalkFileInfo skips the
// directory's contents entirely. If the function returns filepath.SkipDir when
// invoked on a non-directory file system node, WalkFileInfo skips the remaining
// files in the containing directory.
type WalkFileInfoFunc func(osPathname string, info os.FileInfo, err error) error

// WalkFileInfo walks the file tree rooted at the specified directory, calling
// the specified callback function for each file system node in the tree,
// including root, symbolic links, and other node types. The nodes are walked in
// lexical order, which makes the output deterministic but means that for very
// large directories this function can be inefficient.
//
// This function is often much faster than filepath.Walk because it does not
// need to os.Stat every node it encounters, but rather gets the file system
// node type when it reads the parent directory.
func WalkFileInfo(osDirname string, walkFn WalkFileInfoFunc) error {
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
	queue := []string{""}

	var osRelative string // os-specific relative pathname under directory name

	// As we enumerate over the queue and encounter a directory, its children
	// will be added to the work queue.
	for len(queue) > 0 {
		// Unshift a pathname from the queue (breadth-first traversal of
		// hierarchy)
		osRelative, queue = queue[0], queue[1:]
		osPathname := filepath.Join(osDirname, osRelative)

		// walkFn needs to choose how to handle symbolic links, therefore obtain
		// lstat rather than stat.
		fi, err = os.Lstat(osPathname)
		if err == nil {
			err = walkFn(osPathname, fi, nil)
		} else {
			err = walkFn(osPathname, nil, errors.Wrap(err, "cannot read node"))
		}

		if err != nil {
			if err == filepath.SkipDir {
				if fi.Mode()&os.ModeSymlink > 0 {
					// Resolve symbolic link referent to determine whether node
					// is directory or not.
					fi, err = os.Stat(osPathname)
					if err != nil {
						return errors.Wrap(err, "cannot visit node")
					}
				}
				// If current node is directory, then skip this
				// directory. Otherwise, skip all nodes in the same parent
				// directory.
				if !fi.IsDir() {
					// Consume nodes from queue while they have the same parent
					// as the current node.
					osParent := filepath.Dir(osPathname) + osPathSeparator
					for len(queue) > 0 && strings.HasPrefix(queue[0], osParent) {
						queue = queue[1:] // drop sibling from queue
					}
				}

				continue
			}
			return errors.Wrap(err, "DirWalkFunction") // wrap error returned by walkFn
		}

		if fi.IsDir() {
			osChildrenNames, err := childrenFromDirname(osPathname)
			if err != nil {
				return errors.Wrap(err, "cannot get list of directory children")
			}
			sort.Strings(osChildrenNames)
			for _, osChildName := range osChildrenNames {
				switch osChildName {
				case ".", "..":
					// skip
				default:
					queue = append(queue, filepath.Join(osRelative, osChildName))
				}
			}
		}
	}
	return nil
}

// childrenFromDirname returns a lexicographically sorted list of child
// nodes for the specified directory.
func childrenFromDirname(osDirname string) ([]string, error) {
	fh, err := os.Open(osDirname)
	if err != nil {
		return nil, errors.Wrap(err, "cannot Open")
	}

	osChildrenNames, err := fh.Readdirnames(0) // 0: read names of all children
	if err != nil {
		return nil, errors.Wrap(err, "cannot Readdirnames")
	}

	// Close the file handle to the open directory without masking possible
	// previous error value.
	if er := fh.Close(); err == nil {
		err = errors.Wrap(er, "cannot Close")
	}
	return osChildrenNames, err
}
