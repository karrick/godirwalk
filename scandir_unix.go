// +build !windows

package godirwalk

import (
	"os"
	"syscall"
	"unsafe"
)

// DirectoryScanner is an iterator to enumerate the contents of a directory.
type DirectoryScanner struct {
	scratchBuffer []byte // read directory bytes from file system into this buffer
	workBuffer    []byte // points into scratchBuffer, from which we chunk out directory entries
	osDirname     string
	childName     string
	err           error    // err is the error associated with scanning directory
	statErr       error    // statErr is any error return while attempting to stat an entry
	dh            *os.File // used to close directory after done reading
	de            *Dirent  // most recently decoded directory entry
	sde           *syscall.Dirent
	fd            int // file descriptor used to read entries from directory
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
		scratchBuffer: scratchBuffer,
		osDirname:     osDirname,
		dh:            dh,
		fd:            int(dh.Fd()),
	}
	return scanner, nil
}

// done is called when directory scanner unable to continue, with either the
// triggering error, or nil when there are simply no more entries to read from
// the directory.
func (s *DirectoryScanner) done(err error) {
	cerr := s.dh.Close()
	s.dh = nil

	if err == nil {
		s.err = cerr
	} else {
		s.err = err
	}

	s.scratchBuffer, s.workBuffer = nil, nil
	s.osDirname = ""
	s.childName = ""
	s.statErr = nil
	s.de = nil
	s.sde = nil
	s.fd = 0
}

func (s *DirectoryScanner) Dirent() (*Dirent, error) {
	if s.de == nil {
		s.de = &Dirent{name: s.childName}
		s.de.modeType, s.statErr = modeType(s.sde, s.osDirname, s.childName)
	}
	return s.de, s.statErr
}

func (s *DirectoryScanner) Err() error { return s.err }

func (s *DirectoryScanner) Name() string { return s.childName }

// Scan potentially reads and then decodes the next directory entry from the
// file system.
//
// When it returns false, this releases resources used by the DirectoryScanner
// then returns any error associated with closing the file system directory
// resource.
func (s *DirectoryScanner) Scan() bool {
	if s.err != nil {
		return false
	}

	for {
		// When the work buffer has nothing remaining to decode, we need to load
		// more data from disk.
		if len(s.workBuffer) == 0 {
			n, err := syscall.ReadDirent(s.fd, s.scratchBuffer)
			if err != nil {
				s.done(err)
				return false
			}
			if n <= 0 { // end of directory
				s.done(nil)
				return false
			}
			s.workBuffer = s.scratchBuffer[:n] // trim work buffer to number of bytes read
		}

		// Loop until we have a usable file system entry, or we run out of data
		// in the work buffer.
		for len(s.workBuffer) > 0 {
			s.sde = (*syscall.Dirent)(unsafe.Pointer(&s.workBuffer[0])) // point entry to first syscall.Dirent in buffer
			s.workBuffer = s.workBuffer[s.sde.Reclen:]                  // advance buffer for next iteration through loop

			if inoFromDirent(s.sde) == 0 {
				continue // inode set to 0 indicates an entry that was marked as deleted
			}

			nameSlice := nameFromDirent(s.sde)
			namlen := len(nameSlice)
			if (namlen == 0) || (namlen == 1 && nameSlice[0] == '.') || (namlen == 2 && nameSlice[0] == '.' && nameSlice[1] == '.') {
				continue // skip unimportant entries
			}

			s.de = nil
			s.childName = string(nameSlice)
			return true
		}
		// No more data in the work buffer, so loop around in the outside loop
		// to fetch more data.
	}
}
