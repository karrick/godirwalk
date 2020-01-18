// +build !windows

package godirwalk

import (
	"os"
	"syscall"
	"unsafe"
)

func readDirents(osDirname string, scratchBuffer []byte) ([]*Dirent, error) {
	var entries []*Dirent
	var workBuffer []byte

	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}
	fd := int(dh.Fd())

	if len(scratchBuffer) < MinimumScratchBufferSize {
		scratchBuffer = make([]byte, MinimumScratchBufferSize)
	}

reloadWorkBuffer:
	n, err := syscall.ReadDirent(fd, scratchBuffer)
	// n, err := unix.ReadDirent(fd, scratchBuffer)
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

getNextEntry:
	// Loop until we have a usable file system entry, or we run out of data
	// in the work buffer.
	sde := (*syscall.Dirent)(unsafe.Pointer(&workBuffer[0])) // point entry to first syscall.Dirent in buffer
	workBuffer = workBuffer[reclen(sde):] // advance buffer for next iteration through loop

	if inoFromDirent(sde) == 0 {
		goto getNextEntry // inode set to 0 indicates an entry that was marked as deleted
	}

	nameSlice := nameFromDirent(sde)
	nameLength := len(nameSlice)

	if nameLength == 0 || (nameSlice[0] == '.' && (nameLength == 1 || (nameLength == 2 && nameSlice[1] == '.'))) {
		goto getNextEntry
	}

	childName := string(nameSlice)
	mt, err := modeTypeFromDirent(sde, osDirname, childName)
	if err != nil {
		_ = dh.Close()
		return nil, err
	}
	entries = append(entries, &Dirent{name: childName, modeType: mt})

	if len(workBuffer) > 0 {
		goto getNextEntry
	}

	// No more data in the work buffer, so loop around in the outside loop
	// to fetch more data.
	goto reloadWorkBuffer
}

func readDirnames(osDirname string, scratchBuffer []byte) ([]string, error) {
	var entries []string
	var workBuffer []byte
	var sde *syscall.Dirent

	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, err
	}
	fd := int(dh.Fd())

	if len(scratchBuffer) < MinimumScratchBufferSize {
		scratchBuffer = make([]byte, MinimumScratchBufferSize)
	}

reloadWorkBuffer:
	n, err := syscall.ReadDirent(fd, scratchBuffer)
	// n, err := unix.ReadDirent(fd, scratchBuffer)
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

getNextEntry:
	// Loop until we have a usable file system entry, or we run out of data
	// in the work buffer.
	sde = (*syscall.Dirent)(unsafe.Pointer(&workBuffer[0])) // point entry to first syscall.Dirent in buffer
	workBuffer = workBuffer[reclen(sde):] // advance buffer for next iteration through loop

	if inoFromDirent(sde) == 0 {
		goto getNextEntry // inode set to 0 indicates an entry that was marked as deleted
	}

	nameSlice := nameFromDirent(sde)
	nameLength := len(nameSlice)

	if nameLength == 0 || (nameSlice[0] == '.' && (nameLength == 1 || (nameLength == 2 && nameSlice[1] == '.'))) {
		goto getNextEntry
	}

	entries = append(entries, string(nameSlice))

	if len(workBuffer) > 0 {
		goto getNextEntry
	}

	// No more data in the work buffer, so loop around in the outside loop
	// to fetch more data.
	goto reloadWorkBuffer
}
