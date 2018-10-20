// +build !windows

package godirwalk

import (
	"os"

	"github.com/pkg/errors"
)

// symlinkDirHelper accepts a pathname and Dirent for a symlink, and determine
// what the ModeType bits are for what the symlink ultimately points to (such
// as whether it is a directory or not)
func symlinkDirHelper(pathname string, de *Dirent, o *Options) (skip bool, err error) {
	// Only need to Stat entry if platform did not already have os.ModeDir
	// set, such as would be the case for unix like operating systems.
	// (This guard eliminates extra os.Stat check on Windows.)
	if !de.IsDir() {
		fi, stErr := os.Stat(pathname)
		if stErr != nil {
			skip = true
			err = errors.Wrap(stErr, "cannot Stat")
			if action := o.ErrorCallback(pathname, err); action == SkipNode {
				err = nil
				return
			}
			return
		}
		de.modeType = fi.Mode() & os.ModeType
	}
	return
}
