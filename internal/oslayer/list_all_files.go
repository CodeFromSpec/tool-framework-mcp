// code-from-spec: SPEC/golang/implementation/oslayer/list_all_files@8V4nYGpT8xX_tUv-cND1MSiynS4
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

	var results []CfsPath

	walkErr := filepath.WalkDir(string(osPath), func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		converted, convErr := OsPathToCfs(OsPath(entryPath))
		if convErr != nil {
			return convErr
		}
		results = append(results, converted)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("%w: %s", ErrWalkError, walkErr)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})

	return results, nil
}
