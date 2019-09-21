// +build windows

package godirwalk

import (
	"fmt"
	"os"
)

// DirectoryScanner is an iterator to enumerate the contents of a directory.
type DirectoryScanner struct {
	dirent    Dirent
	osDirname string
	dh        *os.File // dh is handle to open directory
	statErr   error    // statErr is any error return while attempting to stat an entry
	err       error    // err is the error associated with scanning directory
}

// NewDirectoryScanner returns a new DirectoryScanner.
func NewDirectoryScanner(osDirname string, _ []byte) (*DirectoryScanner, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}
	scanner := &DirectoryScanner{
		osDirname: osDirname,
		dh:        dh,
	}
	return scanner, nil
}

// Close releases resources used by the DirectoryScanner then returns any error
// associated with closing the file system directory resource.
func (s *DirectoryScanner) Close() error {
	err := s.dh.Close()
	s.dirent.reset()
	s.statErr = nil
	s.osDirname, s.dh, s.err = "", nil, nil
	return err
}

func (s *DirectoryScanner) Dirent() (*Dirent, error) { return &s.dirent, s.statErr }

func (s *DirectoryScanner) Err() error { return s.err }

// Scan potentially reads and then decodes the next directory entry from the
// file system.
func (s *DirectoryScanner) Scan() bool {
	if s.err != nil {
		return false
	}

	fileinfos, err := s.dh.Readdir(1)
	if err != nil {
		s.err = err
		return false
	}

	if l := len(fileinfos); l != 1 {
		s.err = fmt.Errorf("expected a single entry rather than %d", l)
		return false
	}

	fi := fileinfos[0]
	s.dirent.name = fi.Name()
	s.dirent.modeType = fi.Mode() & os.ModeType
	// s.dirent.err = nil
	return true
}
