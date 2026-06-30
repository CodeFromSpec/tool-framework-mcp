// code-from-spec: SPEC/golang/implementation/oslayer/list_all_files@mjaoxI-lx5-aEUDmHgc-GdcAmG8
package oslayer

import (
	"errors"
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

	var resultsListAccumulator []CfsPath

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

		resultsListAccumulator = append(resultsListAccumulator, converted)
		return nil
	})

	if walkErr != nil {
		if errors.Is(walkErr, ErrSymlinkNotAllowed) {
			return nil, walkErr
		}
		if errors.Is(walkErr, ErrWalkError) {
			return nil, walkErr
		}
		return nil, walkErr
	}

	sort.Slice(resultsListAccumulator, func(i, j int) bool {
		return resultsListAccumulator[i] < resultsListAccumulator[j]
	})

	return resultsListAccumulator, nil
}
