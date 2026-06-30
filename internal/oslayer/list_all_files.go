// code-from-spec: SPEC/golang/implementation/oslayer/list_all_files@e1nHxkYhJi8TwsAI-IywSBv_h44
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

	walkErr := filepath.WalkDir(string(osPath), func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%w: %s", ErrSymlinkNotAllowed, entryPath)
		}
		converted, convErr := OsPathToCfs(OsPath(entryPath))
		if convErr != nil {
			return convErr
		}
		resultsList = append(resultsList, converted)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("%w: %s", ErrWalkError, walkErr)
	}

	sort.Slice(resultsList, func(i, j int) bool {
		return resultsList[i] < resultsList[j]
	})

	return resultsList, nil
}
