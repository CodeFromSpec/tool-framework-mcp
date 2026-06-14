// code-from-spec: ROOT/golang/implementation/os/path_utils@woTTwkNmuzmAO69YqPlstZvLuSg
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
var ErrPathAbsolute = errors.New("path must be relative (no leading slash or drive letter)")
var ErrPathContainsBackslash = errors.New("path contains backslash characters")
var ErrDirectoryTraversal = errors.New("path contains directory traversal components")
var ErrResolvesOutsideRoot = errors.New("path resolves outside the project root")

func PathGetProjectRoot() (*PathOs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotDetermineRoot, err)
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

	if strings.Contains(value, "\\") {
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

	nativePath := filepath.FromSlash(cfsPath.Value)
	absPath := filepath.Join(root.Value, nativePath)

	_, statErr := os.Stat(absPath)
	if statErr == nil {
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrResolvesOutsideRoot, err)
		}
		if !strings.HasPrefix(resolved, root.Value) {
			return nil, ErrResolvesOutsideRoot
		}
		return &PathOs{Value: resolved}, nil
	}

	return &PathOs{Value: absPath}, nil
}

func PathOsToCfs(osPath *PathOs) (*PathCfs, error) {
	if osPath == nil {
		return nil, ErrPathEmpty
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	resolvedPath := osPath.Value
	_, statErr := os.Stat(osPath.Value)
	if statErr == nil {
		resolved, err := filepath.EvalSymlinks(osPath.Value)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrResolvesOutsideRoot, err)
		}
		resolvedPath = resolved
	}

	rootWithSep := root.Value
	if !strings.HasSuffix(rootWithSep, string(filepath.Separator)) {
		rootWithSep = rootWithSep + string(filepath.Separator)
	}

	if resolvedPath != root.Value && !strings.HasPrefix(resolvedPath, rootWithSep) {
		return nil, ErrResolvesOutsideRoot
	}

	relative := strings.TrimPrefix(resolvedPath, rootWithSep)
	relative = strings.TrimPrefix(relative, string(filepath.Separator))

	cfsValue := filepath.ToSlash(relative)

	return &PathCfs{Value: cfsValue}, nil
}
