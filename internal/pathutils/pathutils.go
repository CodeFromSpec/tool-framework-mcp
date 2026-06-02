// code-from-spec: ROOT/golang/implementation/os/path_utils@9a0fakjwXxM-tc20PeRF_AD8Chc
package pathutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PathCfs struct {
	Value string
}

type PathOs struct {
	Value string
}

var ErrCannotDetermineRoot = errors.New("cannot determine project root")
var ErrPathEmpty = errors.New("path is empty")
var ErrPathAbsolute = errors.New("path is absolute")
var ErrPathContainsBackslash = errors.New("path contains backslash")
var ErrDirectoryTraversal = errors.New("path contains directory traversal")
var ErrResolvesOutsideRoot = errors.New("path resolves outside project root")

func PathGetProjectRoot() (*PathOs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
	}
	return &PathOs{Value: wd}, nil
}

func PathValidateCfs(value string) error {
	if value == "" {
		return ErrPathEmpty
	}

	if strings.HasPrefix(value, "/") {
		return ErrPathAbsolute
	}
	if len(value) >= 2 && value[1] == ':' && ((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) {
		return ErrPathAbsolute
	}

	if strings.Contains(value, `\`) {
		return ErrPathContainsBackslash
	}

	normalized := filepath.Clean(value)
	for _, component := range strings.Split(normalized, string(filepath.Separator)) {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error) {
	if cfsPath == nil {
		return nil, ErrPathEmpty
	}

	if err := PathValidateCfs(cfsPath.Value); err != nil {
		return nil, err
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	osRelative := strings.ReplaceAll(cfsPath.Value, "/", string(filepath.Separator))
	absPath := filepath.Join(root.Value, osRelative)

	_, statErr := os.Stat(absPath)
	if statErr == nil {
		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}
		if !strings.HasPrefix(resolvedPath, root.Value) {
			return nil, ErrResolvesOutsideRoot
		}
		return &PathOs{Value: resolvedPath}, nil
	}

	return &PathOs{Value: absPath}, nil
}

func PathOsToCfs(osPath *PathOs) (*PathCfs, error) {
	if osPath == nil {
		return nil, ErrResolvesOutsideRoot
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	resolvedPath := osPath.Value
	_, statErr := os.Stat(osPath.Value)
	if statErr == nil {
		resolvedPath, err = filepath.EvalSymlinks(osPath.Value)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}
	}

	if !strings.HasPrefix(resolvedPath, root.Value) {
		return nil, ErrResolvesOutsideRoot
	}

	relPath := strings.TrimPrefix(resolvedPath, root.Value)
	relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

	cfsValue := filepath.ToSlash(relPath)

	return &PathCfs{Value: cfsValue}, nil
}
