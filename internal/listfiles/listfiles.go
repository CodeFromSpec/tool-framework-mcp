// code-from-spec: ROOT/golang/implementation/os/list_files@VXWu7-OibVAQl3EroQAomVbW9uc
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

var ErrDirectoryNotFound = errors.New("directory does not exist")
var ErrWalkError = errors.New("filesystem error while traversing directory")

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

	walkErr := filepath.WalkDir(osPath.Value, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %w", ErrWalkError, err)
		}
		if d.IsDir() {
			return nil
		}
		entryOs := &pathutils.PathOs{Value: path}
		entryCfs, err := pathutils.PathOsToCfs(entryOs)
		if err != nil {
			return err
		}
		results = append(results, entryCfs)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	return results, nil
}
