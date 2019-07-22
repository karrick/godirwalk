// +build !windows

package godirwalk

import (
	"os"
	"syscall"
	"unsafe"
)

// DirectoryScanner is an iterator to enumerate the contents of a directory.
type DirectoryScanner struct {
	readBuffer, workBuffer []byte
	osDirname              string
	entry                  Dirent   // most recently decoded directory entry
	dh                     *os.File // file system directory pointer
	err, nerr              error
	fd                     int
}

// NewDirectoryScanner returns a new DirectoryScanner.
func NewDirectoryScanner(osDirname string, scratchBuffer []byte) (*DirectoryScanner, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}
	if len(scratchBuffer) < MinimumScratchBufferSize {
		scratchBuffer = make([]byte, DefaultScratchBufferSize)
	}
	scanner := &DirectoryScanner{
		readBuffer: scratchBuffer,
		osDirname:  osDirname,
		dh:         dh,
		fd:         int(dh.Fd()),
	}
	return scanner, nil
}

// Close releases resources used by the DirectoryScanner then returns any error
// associated with closing the file system directory resource.
func (s *DirectoryScanner) Close() error {
	err := s.dh.Close()
	s.readBuffer, s.workBuffer, s.dh, s.err, s.osDirname = nil, nil, nil, nil, ""
	s.entry.name, s.entry.modeType = "", 0
	return err
}

func (s *DirectoryScanner) Entry() (*Dirent, error) {
	return &s.entry, s.nerr
}

func (s *DirectoryScanner) Err() error { return s.err }

// Scan potentially reads and then decodes the next directory entry from the
// file system.
func (s *DirectoryScanner) Scan() bool {
	if s.err != nil {
		return false
	}

	for {
		// When the work buffer has nothing remaining to decode, we need to load
		// more data from disk.
		if len(s.workBuffer) == 0 {
			n, err := syscall.ReadDirent(s.fd, s.readBuffer)
			if err != nil {
				s.err = err
				return false
			}
			if n <= 0 {
				return false // end of directory
			}
			s.workBuffer = s.readBuffer[:n]
		}

		// Loop until we have a usable file system entry, or we run out of data
		// in the work buffer.
		for len(s.workBuffer) > 0 {
			de := (*syscall.Dirent)(unsafe.Pointer(&s.workBuffer[0])) // point entry to first syscall.Dirent in buffer
			s.workBuffer = s.workBuffer[de.Reclen:]                   // advance buffer for next iteration through loop

			if inoFromDirent(de) == 0 {
				continue // inode set to 0 indicates an entry that was marked as deleted
			}

			nameSlice := nameFromDirent(de)
			namlen := len(nameSlice)
			if (namlen == 0) || (namlen == 1 && nameSlice[0] == '.') || (namlen == 2 && nameSlice[0] == '.' && nameSlice[1] == '.') {
				continue // skip unimportant entries
			}

			s.entry.name = string(nameSlice)
			s.entry.modeType, s.nerr = modeType(de, s.osDirname, s.entry.name)
			return true
		}
		// No more data in the work buffer, so loop around in the outside loop
		// to fetch more data.
	}
}
