// code-from-spec: ROOT/golang/implementation/os/list_files@qDrfdL6T2r1IidJ2HIIIrEawedQ

package listfiles

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrDirectoryNotFound is returned when the specified directory does not exist.
var ErrDirectoryNotFound = errors.New("directory not found")

// ErrWalkError is returned when a filesystem error occurs while traversing
// the directory tree.
var ErrWalkError = errors.New("filesystem walk error")

// ListFiles returns all files (not directories) found recursively under the
// directory identified by cfsPath. Results are returned as PathCfs values
// sorted alphabetically. If the directory exists but contains no files, an
// empty slice is returned.
//
// Errors:
//   - ErrDirectoryNotFound: the directory does not exist.
//   - ErrWalkError: a filesystem error occurred while traversing the tree.
//   - (PathUtils.*): propagated from pathutils.PathCfsToOs.
//   - (PathUtils.*): propagated from pathutils.PathOsToCfs.
func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error) {
	// Step 1: Convert CFS path to OS path.
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("converting cfs path to os path: %w", err)
	}

	// Step 2: Check that the path refers to an existing directory.
	info, err := os.Stat(osPath.Value)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, cfsPath.Value)
		}
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, cfsPath.Value)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, cfsPath.Value)
	}

	// Step 3: Initialize results slice.
	var results []*pathutils.PathCfs

	// Step 4: Recursively walk the directory tree.
	walkErr := filepath.WalkDir(osPath.Value, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", ErrWalkError, err.Error())
		}

		// Skip directories — continue traversing their contents.
		if d.IsDir() {
			return nil
		}

		// Convert the file's OS path to a CFS path.
		entryOsPath := &pathutils.PathOs{Value: path}
		entryCfsPath, err := pathutils.PathOsToCfs(entryOsPath)
		if err != nil {
			return fmt.Errorf("converting os path to cfs path: %w", err)
		}

		results = append(results, entryCfsPath)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	// Step 5: Sort results alphabetically by their value field.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	// Step 6: Return results.
	return results, nil
}
