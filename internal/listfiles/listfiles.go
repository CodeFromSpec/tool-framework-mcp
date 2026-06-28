// code-from-spec: SPEC/golang/implementation/os/list_files@ZaPhj6GsLDUiaUZayDpQnI5GEK8
package listfiles

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrDirectoryNotFound = errors.New("directory not found")
var ErrWalkError = errors.New("filesystem error occurred while traversing")

func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error) {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(osPath.Value)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, osPath.Value)
	}

	var results []*pathutils.PathCfs

	walkErr := filepath.WalkDir(osPath.Value, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		entryOsPath := &pathutils.PathOs{Value: entryPath}
		entryCfsPath, convErr := pathutils.PathOsToCfs(entryOsPath)
		if convErr != nil {
			return convErr
		}
		results = append(results, entryCfsPath)
		return nil
	})

	if walkErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrWalkError, walkErr)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	return results, nil
}
