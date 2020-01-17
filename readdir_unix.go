// +build !windows

package godirwalk

import (
	"os"
	"syscall"
	"unsafe"
)

func readDirents(osDirname string, scratchBuffer []byte) ([]*Dirent, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}
	fd := int(dh.Fd())

	if len(scratchBuffer) < MinimumScratchBufferSize {
		scratchBuffer = make([]byte, MinimumScratchBufferSize)
	}

	var entries []*Dirent
	var workBuffer []byte
	var sde *syscall.Dirent

	for {
		// When the work buffer has nothing remaining to decode, we need to load
		// more data from disk.
		if len(workBuffer) == 0 {
			n, err := syscall.ReadDirent(fd, scratchBuffer)
			if err != nil {
				_ = dh.Close()
				return nil, err
			}
			if n <= 0 { // end of directory: normal exit
				if err = dh.Close(); err != nil {
					return nil, err
				}
				return entries, nil
			}
			workBuffer = scratchBuffer[:n] // trim work buffer to number of bytes read
		}

		// Loop until we have a usable file system entry, or we run out of data
		// in the work buffer.
		for len(workBuffer) > 0 {
			sde = (*syscall.Dirent)(unsafe.Pointer(&workBuffer[0])) // point entry to first syscall.Dirent in buffer
			workBuffer = workBuffer[reclen(sde):]                   // advance buffer for next iteration through loop

			if inoFromDirent(sde) == 0 {
				continue // inode set to 0 indicates an entry that was marked as deleted
			}

			nameSlice := nameFromDirent(sde)
			nameLength := len(nameSlice)

			if nameLength == 0 || (nameSlice[0] == '.' && (nameLength == 1 || (nameLength == 2 && nameSlice[1] == '.'))) {
				continue
			}

			childName := string(nameSlice)
			mt, err := modeTypeFromDirent(sde, osDirname, childName)
			if err != nil {
				_ = dh.Close()
				return nil, err
			}
			entries = append(entries, &Dirent{name: childName, modeType: mt})
		}
		// No more data in the work buffer, so loop around in the outside loop
		// to fetch more data.
	}

	panic("should not get here")
}

func readDirnames(osDirname string, scratchBuffer []byte) ([]string, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}
	fd := int(dh.Fd())

	if len(scratchBuffer) < MinimumScratchBufferSize {
		scratchBuffer = make([]byte, MinimumScratchBufferSize)
	}

	var entries []string
	var workBuffer []byte
	var sde *syscall.Dirent

	for {
		// When the work buffer has nothing remaining to decode, we need to load
		// more data from disk.
		if len(workBuffer) == 0 {
			n, err := syscall.ReadDirent(fd, scratchBuffer)
			if err != nil {
				_ = dh.Close()
				return nil, err
			}
			if n <= 0 { // end of directory: normal exit
				if err = dh.Close(); err != nil {
					return nil, err
				}
				return entries, nil
			}
			workBuffer = scratchBuffer[:n] // trim work buffer to number of bytes read
		}

		// Loop until we have a usable file system entry, or we run out of data
		// in the work buffer.
		for len(workBuffer) > 0 {
			sde = (*syscall.Dirent)(unsafe.Pointer(&workBuffer[0])) // point entry to first syscall.Dirent in buffer
			workBuffer = workBuffer[reclen(sde):]                   // advance buffer for next iteration through loop

			if inoFromDirent(sde) == 0 {
				continue // inode set to 0 indicates an entry that was marked as deleted
			}

			nameSlice := nameFromDirent(sde)
			nameLength := len(nameSlice)

			if nameLength == 0 || (nameSlice[0] == '.' && (nameLength == 1 || (nameLength == 2 && nameSlice[1] == '.'))) {
				continue
			}

			entries = append(entries, string(nameSlice))
		}
		// No more data in the work buffer, so loop around in the outside loop
		// to fetch more data.
	}

	panic("should not get here")
}
