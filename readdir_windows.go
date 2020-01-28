// +build windows

package godirwalk

// MinimumScratchBufferSize specifies the minimum size of the scratch buffer
// that ReadDirents, ReadDirnames, Scanner, and Walk will use when reading file
// entries from the operating system. During program startup it is initialized
// to the result from calling `os.Getpagesize()` for non Windows environments,
// and 0 for Windows.
var MinimumScratchBufferSize = 0

func newScratchBuffer() []byte { return nil }

func readDirents(osDirname string, _ []byte) (Dirents, error) {
	var entries Dirents
	scanner, err := NewScanner(osDirname)
	if err != nil {
		return nil, err
	}
	for scanner.Scan() {
		if dirent, err := scanner.Dirent(); err == nil {
			entries = append(entries, dirent)
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func readDirnames(osDirname string, _ []byte) ([]string, error) {
	var entries []string
	scanner, err := NewScanner(osDirname)
	if err != nil {
		return nil, err
	}
	for scanner.Scan() {
		entries = append(entries, scanner.Name())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
