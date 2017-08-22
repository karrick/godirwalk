package godirwalk

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

const osPathSeparator = string(filepath.Separator)

// Dirent stores the name and file system mode type of discovered file system
// entries.
type Dirent struct {
	// Name is the filename of the file system entry, relative to its parent.
	Name string

	// ModeType is the mode bits that specify the file system entry type. We
	// could make our own enum-like data type for encoding the file type, but
	// Go's runtime already gives us architecture independent file modes, as
	// discussed in `os/types.go`:
	//
	//    Go's runtime FileMode type has same definition on all systems, so
	//    that information about files can be moved from one system to
	//    another portably.
	ModeType os.FileMode
}

// Dirents represents a slice of direntType pointers, which are sortable by
// name. This type satisfies the `sort.Sortable` interface.
type Dirents []*Dirent

// Len returns the count of Dirent structures in the slice.
func (l Dirents) Len() int { return len(l) }

// Less returns true if and only if the Name of the element specified by the
// first index is lexicographically less than that of the second index.
func (l Dirents) Less(i, j int) bool { return l[i].Name < l[j].Name }

// Swap exchanges the two Dirent entries specified by the two provided indexes.
func (l Dirents) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// GetDirectoryEntries returns a slice of pointers to Dirent structures,
// representing the file system children of the specified directory. If the
// specified directory is a symbolic link, it will be resolved.
func GetDirectoryEntries(osDirname string) (Dirents, error) {
	return GetDirectoryEntriesBuffer(osDirname, make([]byte, bufsize))
}

const bufsize = 16 * 1024

// GetDirectoryEntriesBuffer returns a slice of pointers to Dirent structures,
// representing the file system children of the specified directory. If the
// specified directory is a symbolic link, it will be resolved. If the optional
// scratch buffer is provided, it will be used to avoid excess allocations;
// otherwise, a one-time scratch buffer will be allocated and released when this
// function exits.
func GetDirectoryEntriesBuffer(osDirname string, optionalScratchBuffer []byte) (Dirents, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open directory")
	}
	dfd := int(dh.Fd())

	var entries Dirents

	if optionalScratchBuffer == nil {
		optionalScratchBuffer = make([]byte, 16*1024)
	}

	var nameBytes []byte                                     // will be updated to point to syscall.Dirent.Name
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&nameBytes)) // save slice header, so we can re-use each loop

	for {
		n, err := syscall.ReadDirent(dfd, optionalScratchBuffer)
		if err != nil {
			_ = dh.Close() // ignore possible error returned by Close
			return nil, errors.Wrap(err, "cannot read directory")
		}
		if n <= 0 {
			break // end of directory reached
		}
		// Loop over the bytes returned by reading the directory entries.
		buf := optionalScratchBuffer[:n]
		for len(buf) > 0 {
			// unshift left-most directory entry from the buffer
			de := (*syscall.Dirent)(unsafe.Pointer(&buf[0]))
			buf = buf[de.Reclen:]

			if de.Ino == 0 {
				continue // this item has been deleted, but not yet removed from directory
			}

			// Convert syscall.Dirent.Name, which is array of int8, to []byte,
			// by overwriting Cap, Len, and Data slice members to values from
			// syscall.Dirent.
			//
			// ??? Set the upper bound on the Cap and the Len to the size of the
			// record length of the dirent.
			sh.Cap, sh.Len, sh.Data = int(de.Reclen), int(de.Reclen), uintptr(unsafe.Pointer(&de.Name[0]))
			namlen := bytes.IndexByte(nameBytes, 0) // look for NULL byte
			if namlen == -1 {
				namlen = len(de.Name)
			}
			nameBytes = nameBytes[:namlen]

			// Skip "." and ".." entries.
			if namlen == 1 && nameBytes[0] == '.' || namlen == 2 && nameBytes[0] == '.' && nameBytes[1] == '.' {
				continue
			}

			nameString := string(nameBytes)

			// Convert syscall constant, which is in purview of OS, to a
			// constant defined by Go, assumed by this project to be stable.
			var mode os.FileMode
			switch de.Type {
			case syscall.DT_REG:
				// regular file
			case syscall.DT_DIR:
				mode = os.ModeDir
			case syscall.DT_LNK:
				mode = os.ModeSymlink
			case syscall.DT_BLK:
				mode = os.ModeDevice
			case syscall.DT_CHR:
				mode = os.ModeDevice | os.ModeCharDevice
			case syscall.DT_FIFO:
				mode = os.ModeNamedPipe
			case syscall.DT_SOCK:
				mode = os.ModeSocket
			default:
				// If syscall returned unknown type (e.g., DT_UNKNOWN, DT_WHT),
				// then resolve actual mode by getting stat.
				fi, err := os.Stat(filepath.Join(osDirname, nameString))
				if err != nil {
					_ = dh.Close() // ignore possible error returned by Close
					return nil, errors.Wrap(err, "cannot stat")
				}
				mode = fi.Mode()
			}

			// We only care about the bits that identify the type of a file
			// system node, and can ignore append, exclusive, temporary, setuid,
			// setgid, permission bits, and sticky bits, which are coincident to
			// bits which declare type of the file system node.
			entries = append(entries, &Dirent{Name: nameString, ModeType: mode & os.ModeType})
		}
	}
	if err = dh.Close(); err != nil {
		return nil, err
	}
	return entries, nil
}
