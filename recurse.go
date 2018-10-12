package godirwalk

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

func init() {
	MinimumScratchBufferSize = os.Getpagesize()
}

// Options provide parameters for how the Walk function operates.
type RecurseOptions struct {
	// ErrorCallback specifies a function to be invoked in the case of an error
	// that could potentially be ignored while walking a file system
	// hierarchy. When set to nil or left as its zero-value, any error condition
	// causes Walk to immediately return the error describing what took
	// place. When non-nil, this user supplied function is invoked with the OS
	// pathname of the file system object that caused the error along with the
	// error that took place. The return value of the supplied ErrorCallback
	// function determines whether the error will cause Walk to halt immediately
	// as it would were no ErrorCallback value provided, or skip this file
	// system node yet continue on with the remaining nodes in the file system
	// hierarchy.
	//
	// ErrorCallback is invoked both for errors that are returned by the
	// runtime, and for errors returned by other user supplied callback
	// functions.
	ErrorCallback func(string, error) ErrorAction

	// FollowSymbolicLinks specifies whether Walk will follow symbolic links
	// that refer to directories. When set to false or left as its zero-value,
	// Walk will still invoke the callback function with symbolic link nodes,
	// but if the symbolic link refers to a directory, it will not recurse on
	// that directory. When set to true, Walk will recurse on symbolic links
	// that refer to a directory.
	FollowSymbolicLinks bool

	// Unsorted controls whether or not Walk will sort the immediate descendants
	// of a directory by their relative names prior to visiting each of those
	// entries.
	//
	// When set to false or left at its zero-value, Walk will get the list of
	// immediate descendants of a particular directory, sort that list by
	// lexical order of their names, and then visit each node in the list in
	// sorted order. This will cause Walk to always traverse the same directory
	// tree in the same order, however may be inefficient for directories with
	// many immediate descendants.
	//
	// When set to true, Walk skips sorting the list of immediate descendants
	// for a directory, and simply visits each node in the order the operating
	// system enumerated them. This will be more fast, but with the side effect
	// that the traversal order may be different from one invocation to the
	// next.
	Unsorted bool

	// Callback is a required function that Walk will invoke for every file
	// system node it encounters.
	Callback RecurseFunc

	// PostChildrenCallback is an option function that Recurse will invoke for
	// every file system directory it encounters before its children are
	// processed. If the function returns filepath.SkipDir, then the children
	// and the folder are not processed.
	PreChildrenCallback WalkFunc

	// ScratchBuffer is an optional byte slice to use as a scratch buffer for
	// Walk to use when reading directory entries, to reduce amount of garbage
	// generation. Not all architectures take advantage of the scratch
	// buffer. If omitted or the provided buffer has fewer bytes than
	// MinimumScratchBufferSize, then a buffer with DefaultScratchBufferSize
	// bytes will be created and used once per Walk invocation.
	ScratchBuffer []byte
}

// RecurseFunc is the type of the function called for each file system node 
// visited by Recurse. The pathname argument will contain the argument to Walk as
// a prefix; that is, if Recurse is called with "dir", which is a directory 
// containing the file "a", the provided WalkFunc will be invoked with the 
// argument "dir/a", using the correct os.PathSeparator for the Go Operating 
// System architecture, GOOS. The directory entry argument is a pointer to a 
// Dirent for the node, providing access to both the basename and the mode type 
// of the file system node. The sibling and child RecurseResult are results of
// Callback called respectively on previously visited files in the same 
// directory and on evry children of directory if it is a directory.
//
// If your directory structure looks like :
// a/
// a/b
// a/c
// a/d
// a/d/g
// Then the call stack will be :
// r(a/, s, r(a/d, r(a/c, r(a/b, s, s), s), r(a/d/g, s, s)))
// where r is the recurse function and s is the startValue passed to 
// godirwalk.Recurse function.
//
// If an error is returned by the Callback or PreChildrenCallback functions,
// and no ErrorCallback function is provided, processing stops. If an
// ErrorCallback function is provided, then it is invoked with the OS pathname
// of the node that caused the error along along with the error. The return
// value of the ErrorCallback function determines whether to halt processing, or
// skip this node and continue processing remaining file system nodes.
//
// The exception is when the function returns the special value
// filepath.SkipDir. If the function returns filepath.SkipDir when invoked on a
// directory, Walk skips the directory's contents entirely. If the function
// returns filepath.SkipDir when invoked on a non-directory file system node,
// Walk skips the remaining files in the containing directory. Note that any
// supplied ErrorCallback function is not invoked with filepath.SkipDir when the
// Callback or PostChildrenCallback functions return that special value.
type RecurseFunc func(osPathname string, directoryEntry *Dirent, sibling, child int64) (int64, error)

// RecurseResult is empty interface, allowing Callbacks to work with any 
// type of data (well ... it should)
type RecurseResult interface {}

// Recurse walks the file tree rooted at the specified directory, calling the
// specified callback function for each file system node in the tree, including
// root, symbolic links, and other node types. The nodes are walked in lexical
// order, which makes the output deterministic but means that for very large
// directories this function can be inefficient.
//
// This function is often much faster than filepath.Walk because it does not
// invoke os.Stat for every node it encounters, but rather obtains the file
// system node type when it reads the parent directory.
//
// If a runtime error occurs, either from the operating system or from the
// upstream Callback or PreChildrenCallback functions, processing typically
// halts. However, when an ErrorCallback function is provided in the provided
// Options structure, that function is invoked with the error along with the OS
// pathname of the file system node that caused the error. The ErrorCallback
// function's return value determines the action that Walk will then take.
//
//    func main() {
//        dirname := "."
//        if len(os.Args) > 1 {
//            dirname = os.Args[1]
//        }
//	_, err := godirwalk.Recurse(dirname, &godirwalk.RecurseOptions{
//		Callback: func(osPathname string, de *godirwalk.Dirent, siblingSize, childSize int64) (int64, error) {
//			if de.IsDir() {
//				fmt.Printf("%d	%s\n", childSize, osPathname)
//			}
//			return siblingSize+childSize+1, nil
//		},
//		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
//			// Your program may want to log the error somehow.
//			// fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
//
//			// For the purposes of this example, a simple SkipNode will suffice,
//			// although in reality perhaps additional logic might be called for.
//			return godirwalk.SkipNode
//		},
//		Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
//	}, int64(0))
//        if err != nil {
//            fmt.Fprintf(os.Stderr, "%s\n", err)
//            os.Exit(1)
//        }
//    }
func Recurse(pathname string, options *RecurseOptions, startValue int64) (int64, error) {
	pathname = filepath.Clean(pathname)

	var fi os.FileInfo
	var err error

	if options.FollowSymbolicLinks {
		fi, err = os.Stat(pathname)
		if err != nil {
			return startValue, errors.Wrap(err, "cannot Stat")
		}
	} else {
		fi, err = os.Lstat(pathname)
		if err != nil {
			return startValue, errors.Wrap(err, "cannot Lstat")
		}
	}

	mode := fi.Mode()
	if mode&os.ModeDir == 0 {
		return startValue, errors.Errorf("cannot Recurse non-directory: %s", pathname)
	}

	dirent := &Dirent{
		name:     filepath.Base(pathname),
		modeType: mode & os.ModeType,
	}

	// If ErrorCallback is nil, set to a default value that halts the walk
	// process on all operating system errors. This is done to allow error
	// handling to be more succinct in the walk code.
	if options.ErrorCallback == nil {
		options.ErrorCallback = defaultErrorCallback
	}

	if len(options.ScratchBuffer) < MinimumScratchBufferSize {
		options.ScratchBuffer = make([]byte, DefaultScratchBufferSize)
	}

	result, err := recurse(pathname, dirent, options, startValue, startValue)
	if err == filepath.SkipDir {
		return result, nil // silence SkipDir for top level
	}
	return result, err
}

// walk recursively traverses the file system node specified by pathname and the
// Dirent.
func recurse(osPathname string, dirent *Dirent, options *RecurseOptions, siblingValue, startValue int64) (int64, error) {
	// On some platforms, an entry can have more than one mode type bit set.
	// For instance, it could have both the symlink bit and the directory bit
	// set indicating it's a symlink to a directory.
	if dirent.IsSymlink() {
		if !options.FollowSymbolicLinks {
			return siblingValue, nil
		}
		// Only need to Stat entry if platform did not already have os.ModeDir
		// set, such as would be the case for unix like operating systems. (This
		// guard eliminates extra os.Stat check on Windows.)
		if !dirent.IsDir() {
			referent, err := os.Readlink(osPathname)
			if err != nil {
				err = errors.Wrap(err, "cannot Readlink")
				if action := options.ErrorCallback(osPathname, err); action == SkipNode {
					return siblingValue, nil
				}
				return siblingValue, err
			}

			var osp string
			if filepath.IsAbs(referent) {
				osp = referent
			} else {
				osp = filepath.Join(filepath.Dir(osPathname), referent)
			}

			fi, err := os.Stat(osp)
			if err != nil {
				err = errors.Wrap(err, "cannot Stat")
				if action := options.ErrorCallback(osp, err); action == SkipNode {
					return siblingValue, nil
				}
				return siblingValue, err
			}
			dirent.modeType = fi.Mode() & os.ModeType
		}
	}

	if options.PreChildrenCallback != nil {
		err := options.PreChildrenCallback(osPathname, dirent)
		if err == filepath.SkipDir {
			return siblingValue, err
		}

		if err != nil {
			err = errors.Wrap(err, "PreChildrenCallback") // wrap potential errors returned by callback
			if action := options.ErrorCallback(osPathname, err); action == SkipNode {
				return siblingValue, nil
			} else {
				return siblingValue, err
			}
		}
	}

	childValue := startValue
	if dirent.IsDir() {

		deChildren, err := ReadDirents(osPathname, options.ScratchBuffer)
		if err != nil {
			err = errors.Wrap(err, "cannot ReadDirents")
			if action := options.ErrorCallback(osPathname, err); action == SkipNode {
				return siblingValue, nil
			}
			return siblingValue, err
		}

		if !options.Unsorted {
			sort.Sort(deChildren) // sort children entries unless upstream says to leave unsorted
		}

		for _, deChild := range deChildren {
			osChildname := filepath.Join(osPathname, deChild.name)
			childValue, err = recurse(osChildname, deChild, options, childValue, startValue)
			if err != nil {
				if err != filepath.SkipDir {
					return siblingValue, err
				}
				// If received skipdir on a directory, stop processing that
				// directory, but continue to its siblings. If received skipdir on a
				// non-directory, stop processing remaining siblings.
				if deChild.IsSymlink() {
					// Only need to Stat entry if platform did not already have
					// os.ModeDir set, such as would be the case for unix like
					// operating systems. (This guard eliminates extra os.Stat check
					// on Windows.)
					if !deChild.IsDir() {
						// Resolve symbolic link referent to determine whether node
						// is directory or not.
						referent, err := os.Readlink(osChildname)
						if err != nil {
							err = errors.Wrap(err, "cannot Readlink")
							if action := options.ErrorCallback(osChildname, err); action == SkipNode {
								continue // with next child
							}
							return siblingValue, err
						}

						var osp string
						if filepath.IsAbs(referent) {
							osp = referent
						} else {
							osp = filepath.Join(osPathname, referent)
						}

						fi, err := os.Stat(osp)
						if err != nil {
							err = errors.Wrap(err, "cannot Stat")
							if action := options.ErrorCallback(osp, err); action == SkipNode {
								continue // with next child
							}
							return siblingValue, err
						}
						deChild.modeType = fi.Mode() & os.ModeType
					}
				}
				if !deChild.IsDir() {
					// If not directory, return immediately, thus skipping remainder
					// of siblings.
					return siblingValue, nil
				}
			}
		}
	}

	childValue, err := options.Callback(osPathname, dirent, siblingValue, childValue)
	if err != nil {
		if err == filepath.SkipDir {
			return siblingValue, err
		}
		err = errors.Wrap(err, "Callback") // wrap potential errors returned by callback
		if action := options.ErrorCallback(osPathname, err); action == SkipNode {
			return siblingValue, nil
		}
		return siblingValue, err
	}

	return childValue, err
}
