package oslayer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func ListAllFiles(cfsPath CfsPath) ([]CfsPath, error) {
	osPath, err := CfsPathToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(string(osPath))
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, cfsPath)
	}

	var resultsList []CfsPath

	walkErr := filepath.WalkDir(string(osPath), func(entryOsPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", ErrWalkError, entryOsPath)
		}

		if d.IsDir() {
			return nil
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return ErrSymlinkNotAllowed
		}

		converted, convErr := OsPathToCfs(OsPath(entryOsPath))
		if convErr != nil {
			return convErr
		}

		resultsList = append(resultsList, converted)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(resultsList, func(i, j int) bool {
		return resultsList[i] < resultsList[j]
	})

	return resultsList, nil
}
