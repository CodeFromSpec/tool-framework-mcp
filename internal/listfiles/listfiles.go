// code-from-spec: ROOT/golang/implementation/os/list_files@FsD5GCW8xjNQ_GOarwEfuGXGels

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

var (
	// ErrDirectoryNotFound is returned when the given directory does not exist.
	ErrDirectoryNotFound = errors.New("directory not found")

	// ErrWalk is returned when a filesystem error occurs while traversing the directory.
	ErrWalk = errors.New("walk error")
)

// ListFiles returns all files (not directories) found recursively under the
// given directory, as a sorted list of PathCfs values.
//
// If the directory exists but contains no files, an empty list is returned.
//
// Possible errors:
//   - pathutils.ErrPathEmpty
//   - pathutils.ErrPathAbsolute
//   - pathutils.ErrPathContainsBackslash
//   - pathutils.ErrDirectoryTraversal
//   - pathutils.ErrResolvesOutsideRoot
//   - pathutils.ErrCannotDetermineRoot
//   - ErrDirectoryNotFound
//   - ErrWalk
func ListFiles(cfs_path *pathutils.PathCfs) ([]*pathutils.PathCfs, error) {
	// Step 1: Convert the CFS path to an OS path.
	osPath, err := pathutils.PathCfsToOs(cfs_path)
	if err != nil {
		return nil, err
	}

	// Step 2: Check that the path refers to an existing directory.
	info, err := os.Stat(osPath.Value)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, osPath.Value)
		}
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, osPath.Value)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, osPath.Value)
	}

	// Step 3: Create an empty results list.
	results := []*pathutils.PathCfs{}

	// Step 4: Walk the directory recursively.
	walkErr := filepath.WalkDir(osPath.Value, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", ErrWalk, err.Error())
		}

		// Skip directories.
		if d.IsDir() {
			return nil
		}

		// Convert the OS path of the file back to a CFS path.
		entryOsPath := &pathutils.PathOs{Value: entryPath}
		cfsPart, convertErr := pathutils.PathOsToCfs(entryOsPath)
		if convertErr != nil {
			return convertErr
		}

		results = append(results, cfsPart)
		return nil
	})
	if walkErr != nil {
		// If the error is already wrapped with ErrWalk or is a propagated conversion error, return it directly.
		if errors.Is(walkErr, ErrWalk) {
			return nil, walkErr
		}
		return nil, walkErr
	}

	// Step 5: Sort results alphabetically by their value field.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	// Step 6: Return results.
	return results, nil
}
