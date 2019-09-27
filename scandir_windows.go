// +build windows

package godirwalk

import (
	"fmt"
	"os"
)

// Scanner is an iterator to enumerate the contents of a directory.
type Scanner struct {
	osDirname string
	dh        *os.File // dh is handle to open directory
	dirent    *Dirent
	err       error // err is the error associated with scanning directory
}

// NewScanner returns a new directory Scanner.
func NewScanner(osDirname string, _ []byte) (*Scanner, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}
	scanner := &Scanner{
		osDirname: osDirname,
		dh:        dh,
	}
	return scanner, nil
}

// Close releases resources used by the Scanner then returns any error
// associated with closing the file system directory resource.
func (s *Scanner) Close() error {
	err := s.dh.Close()
	s.dirent.reset()
	s.osDirname, s.dh, s.err = "", nil, nil
	return err
}

// Dirent returns the current directory entry while scanning a directory.
func (s *Scanner) Dirent() (*Dirent, error) { return s.dirent, nil }

// Err returns the error associated with scanning a directory.
func (s *Scanner) Err() error { return s.err }

// Name returns the name of the current directory entry while scanning a
// directory.
func (s *Scanner) Name() string { return s.dirent.name }

// Scan potentially reads and then decodes the next directory entry from the
// file system.
func (s *Scanner) Scan() bool {
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
	s.dirent = &Dirent{
		name:     fi.Name(),
		modeType: fi.Mode() & os.ModeType,
	}
	return true
}
