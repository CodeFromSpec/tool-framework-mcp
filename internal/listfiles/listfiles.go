// code-from-spec: ROOT/golang/implementation/os/list_files@qRjmzOHjcCDF8bGi9SOCu4DMRp0
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

var ErrDirectoryNotFound = errors.New("directory not found")
var ErrWalkError = errors.New("filesystem walk error")

func ListFiles(cfs_path *pathutils.PathCfs) ([]*pathutils.PathCfs, error) {
	osPath, err := pathutils.PathCfsToOs(cfs_path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(osPath.Value)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, osPath.Value)
	}

	var results []*pathutils.PathCfs

	walkErr := filepath.WalkDir(osPath.Value, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		cfsEntry, convErr := pathutils.PathOsToCfs(&pathutils.PathOs{Value: path})
		if convErr != nil {
			return convErr
		}
		results = append(results, cfsEntry)
		return nil
	})

	if walkErr != nil {
		return nil, fmt.Errorf("%w: %s", ErrWalkError, walkErr)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	if results == nil {
		results = []*pathutils.PathCfs{}
	}

	return results, nil
}
