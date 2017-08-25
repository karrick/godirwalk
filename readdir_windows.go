package godirwalk

import (
	"os"

	"github.com/pkg/errors"
)

// The functions in this file are mere wrappers of what is already provided by
// standard library, in order to provide the same API as this library provides.
//
// Please send PR or link to article if there is a more performant way of
// enumerating directory contents.

func readdirnames(osDirname string, max int) ([]string, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, errors.Wrap(err, "cannot Open")
	}

	entries, err := dh.Readdirnames(max)
	if er := dh.Close(); err == nil {
		err = er
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot Readdirnames")
	}

	return entries, nil
}

func readdirents(osDirname string, max int) (Dirents, error) {
	dh, err := os.Open(osDirname)
	if err != nil {
		return nil, errors.Wrap(err, "cannot Open")
	}

	fileinfos, err := dh.Readdir(max)
	if er := dh.Close(); err == nil {
		err = er
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot Readdir")
	}

	entries := make(Dirents, len(fileinfos))
	for i, v := range fileinfos {
		entries[i] = &Dirent{name: v.Name(), modeType: v.Mode() & os.ModeType}
	}

	return entries, nil
}
