// code-from-spec: ROOT/golang/implementation/os/path_utils@M2CJLATtYNNoMihwyxXNUuCeaw4
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

var ErrCannotDetermineRoot   = errors.New("cannot determine project root")
var ErrPathEmpty             = errors.New("path is empty")
var ErrPathAbsolute          = errors.New("path must be relative")
var ErrPathContainsBackslash = errors.New("path contains backslash")
var ErrDirectoryTraversal    = errors.New("path contains directory traversal")
var ErrResolvesOutsideRoot   = errors.New("path resolves outside project root")

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

	if len(value) >= 2 && value[1] == ':' {
		return ErrPathAbsolute
	}

	if strings.ContainsRune(value, '\\') {
		return ErrPathContainsBackslash
	}

	normalized := filepath.ToSlash(filepath.Clean(value))
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

func PathCfsToOs(cfs_path *PathCfs) (*PathOs, error) {
	if err := PathValidateCfs(cfs_path.Value); err != nil {
		return nil, err
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	osRelative := filepath.FromSlash(cfs_path.Value)
	absPath := filepath.Join(root.Value, osRelative)

	_, statErr := os.Stat(absPath)
	if statErr == nil {
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks: %w", err)
		}
		if !strings.HasPrefix(resolved, root.Value) {
			return nil, ErrResolvesOutsideRoot
		}
		return &PathOs{Value: resolved}, nil
	}

	return &PathOs{Value: absPath}, nil
}

func PathOsToCfs(os_path *PathOs) (*PathCfs, error) {
	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	resolvedPath := os_path.Value
	_, statErr := os.Stat(os_path.Value)
	if statErr == nil {
		resolvedPath, err = filepath.EvalSymlinks(os_path.Value)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks: %w", err)
		}
	}

	if !strings.HasPrefix(resolvedPath, root.Value) {
		return nil, ErrResolvesOutsideRoot
	}

	relPath := resolvedPath[len(root.Value):]
	relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

	cfsPath := filepath.ToSlash(relPath)

	return &PathCfs{Value: cfsPath}, nil
}
