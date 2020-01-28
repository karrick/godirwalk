// +build windows

package godirwalk

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
