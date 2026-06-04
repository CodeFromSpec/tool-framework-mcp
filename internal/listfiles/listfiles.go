// code-from-spec: ROOT/golang/implementation/os/list_files@t4INp3r6dn2_G0BOg-9SaZUY-Ps
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

func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error) {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(osPath.Value)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrDirectoryNotFound
		}
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, err)
	}
	if !info.IsDir() {
		return nil, ErrDirectoryNotFound
	}

	var results []*pathutils.PathCfs

	walkErr := filepath.WalkDir(osPath.Value, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%w: %s", ErrWalkError, err)
		}
		if d.IsDir() {
			return nil
		}
		cfs, convertErr := pathutils.PathOsToCfs(&pathutils.PathOs{Value: path})
		if convertErr != nil {
			return convertErr
		}
		results = append(results, cfs)
		return nil
	})
	if walkErr != nil {
		if errors.Is(walkErr, ErrWalkError) {
			return nil, walkErr
		}
		return nil, walkErr
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	if results == nil {
		return []*pathutils.PathCfs{}, nil
	}
	return results, nil
}
