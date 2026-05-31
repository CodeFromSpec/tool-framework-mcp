// code-from-spec: ROOT/golang/implementation/os/list_files@giODF5txl5la8V-cE3NAoRbiLcc

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

// ErrDirectoryNotFound is returned when the given directory does not exist.
var ErrDirectoryNotFound = errors.New("directory not found")

// ErrWalkError is returned when a filesystem error occurs while traversing
// the directory tree.
var ErrWalkError = errors.New("walk error")

// ListFiles returns all files (not directories) found recursively under the
// given directory. Results are PathCfs values sorted alphabetically. If the
// directory exists but contains no files, returns an empty slice.
//
// Errors:
//   - ErrDirectoryNotFound: the directory does not exist.
//   - ErrWalkError: a filesystem error occurred while traversing.
//   - (PathUtils.*): propagated from PathCfsToOs.
//   - (PathUtils.*): propagated from PathOsToCfs.
func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error) {
	// Step 1: Convert CFS path to OS-native absolute path.
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("ListFiles: %w", err)
	}

	// Step 2: Check that the path refers to an existing directory.
	info, err := os.Stat(osPath.Value)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ListFiles: %w", ErrDirectoryNotFound)
		}
		return nil, fmt.Errorf("ListFiles: %w", ErrDirectoryNotFound)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("ListFiles: %w", ErrDirectoryNotFound)
	}

	// Step 3 & 4: Walk the directory tree recursively.
	var results []*pathutils.PathCfs

	walkErr := filepath.WalkDir(osPath.Value, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %v", ErrWalkError, err)
		}

		// Step 4: Skip directories, process files.
		if d.IsDir() {
			return nil
		}

		// Step 4a: Convert the file's OS path to a PathCfs.
		fileCfsPath, convertErr := pathutils.PathOsToCfs(&pathutils.PathOs{Value: path})
		if convertErr != nil {
			return convertErr
		}

		// Step 4b: Append the resulting PathCfs to the result list.
		results = append(results, fileCfsPath)
		return nil
	})

	if walkErr != nil {
		// Distinguish between a walk initiation error and a propagated PathOsToCfs error.
		if errors.Is(walkErr, ErrWalkError) {
			return nil, fmt.Errorf("ListFiles: %w", walkErr)
		}
		return nil, fmt.Errorf("ListFiles: %w", walkErr)
	}

	// Step 5: Sort the result list alphabetically by the PathCfs value field.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	// Step 6: Return the sorted list. If no files were found, return an empty slice.
	if results == nil {
		results = []*pathutils.PathCfs{}
	}

	return results, nil
}
