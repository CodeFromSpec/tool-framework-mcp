// code-from-spec: ROOT/golang/implementation/os/list_files@cXSA4EeLkwhSsy1Es-B-Ktnpdxs

package listfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var (
	// ErrDirectoryNotFound is returned when the given directory does not exist.
	ErrDirectoryNotFound = errors.New("directory not found")

	// ErrWalkError is returned when a filesystem error occurs while traversing
	// the directory tree.
	ErrWalkError = errors.New("walk error")
)

// ListFiles returns all files (not directories) found recursively under
// the given directory. Results are PathCfs values sorted alphabetically.
// If the directory exists but contains no files, an empty slice is returned.
//
// Returns an error if:
//   - validation of cfs_path fails (errors propagated from PathCfsToOs).
//   - conversion of discovered OS paths to CFS paths fails (errors
//     propagated from PathOsToCfs).
//   - the directory does not exist (ErrDirectoryNotFound).
//   - a filesystem error occurs while traversing (ErrWalkError).
func ListFiles(cfs_path *pathutils.PathCfs) ([]*pathutils.PathCfs, error) {
	// Step 1: Convert CFS path to OS path.
	osPath, err := pathutils.PathCfsToOs(cfs_path)
	if err != nil {
		return nil, err
	}

	// Step 2: Check that the directory exists.
	info, err := os.Stat(osPath.Value)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, cfs_path.Value)
		}
		return nil, fmt.Errorf("%w: %w", ErrDirectoryNotFound, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s is not a directory", ErrDirectoryNotFound, cfs_path.Value)
	}

	// Step 3: Initialize results slice.
	results := make([]*pathutils.PathCfs, 0)

	// Step 4: Walk the directory recursively.
	walkErr := filepath.WalkDir(osPath.Value, func(path string, d os.DirEntry, err error) error {
		// Filesystem error reading this entry.
		if err != nil {
			return fmt.Errorf("%w: %w", ErrWalkError, err)
		}

		// Skip directories.
		if d.IsDir() {
			return nil
		}

		// Convert the OS path to a CFS path.
		cfsEntry, convErr := pathutils.PathOsToCfs(&pathutils.PathOs{Value: path})
		if convErr != nil {
			return fmt.Errorf("%w: %w", ErrWalkError, convErr)
		}

		results = append(results, cfsEntry)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	// Step 5: Sort results alphabetically by value.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	// Step 6: Return results.
	return results, nil
}
