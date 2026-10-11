package godirwalk

import (
	"errors"
	"os"
	"path/filepath"
)

// WalkParents walks upward through parent directories from osPathname towards
// the filesystem root, invoking callback for each directory. Traversal halts
// when callback returns true, an error occurs, or the filesystem root is
// reached.
//
// When osPathname is relative, WalkParents resolves it to an absolute path.
// When osPathname references an existing non-directory file, traversal begins
// at its containing directory.
func WalkParents(osPathname string, callback func(dir string) (bool, error)) error {
	if callback == nil {
		return errors.New("cannot walk parents with nil callback")
	}

	absPath, err := filepath.Abs(osPathname)
	if err != nil {
		return err
	}

	if fi, err := os.Stat(absPath); err == nil && !fi.IsDir() {
		absPath = filepath.Dir(absPath)
	}

	current := absPath
	for {
		stop, err := callback(current)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return nil
}

// WalkAncestors is an alias for WalkParents.
func WalkAncestors(osPathname string, callback func(dir string) (bool, error)) error {
	return WalkParents(osPathname, callback)
}
