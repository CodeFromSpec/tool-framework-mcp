// code-from-spec: SPEC/golang/implementation/os/path_utils@sb3qy48FGqJ8S0bIqwGN_Prbnww
package pathutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrCannotDetermineRoot = errors.New("cannot determine project root")
var ErrPathEmpty = errors.New("path is empty")
var ErrPathAbsolute = errors.New("path must not be absolute")
var ErrPathContainsBackslash = errors.New("path must not contain backslashes")
var ErrDirectoryTraversal = errors.New("path contains directory traversal components")
var ErrResolvesOutsideRoot = errors.New("path resolves outside the project root")

type PathCfs struct {
	Value string
}

type PathOs struct {
	Value string
}

func PathGetProjectRoot() (PathOs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return PathOs{}, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
	}
	return PathOs{Value: wd}, nil
}

func PathValidateCfs(value string) error {
	if value == "" {
		return ErrPathEmpty
	}

	if strings.HasPrefix(value, "/") || strings.Contains(value, ":") {
		return ErrPathAbsolute
	}

	if strings.Contains(value, "\\") {
		return ErrPathContainsBackslash
	}

	cleaned := filepath.ToSlash(filepath.Clean(value))
	for _, component := range strings.Split(cleaned, "/") {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

func PathCfsToOs(cfsPath PathCfs) (PathOs, error) {
	if err := PathValidateCfs(cfsPath.Value); err != nil {
		return PathOs{}, err
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return PathOs{}, err
	}

	osRelative := filepath.FromSlash(cfsPath.Value)
	absolutePath := filepath.Join(root.Value, osRelative)

	_, statErr := os.Stat(absolutePath)
	if statErr == nil {
		resolved, err := filepath.EvalSymlinks(absolutePath)
		if err != nil {
			return PathOs{}, fmt.Errorf("failed to resolve symlinks: %w", err)
		}
		if !strings.HasPrefix(resolved, root.Value) {
			return PathOs{}, ErrResolvesOutsideRoot
		}
		absolutePath = resolved
	}

	return PathOs{Value: absolutePath}, nil
}

func PathOsToCfs(osPath PathOs) (PathCfs, error) {
	root, err := PathGetProjectRoot()
	if err != nil {
		return PathCfs{}, err
	}

	resolvedValue := osPath.Value
	_, statErr := os.Stat(osPath.Value)
	if statErr == nil {
		resolved, err := filepath.EvalSymlinks(osPath.Value)
		if err != nil {
			return PathCfs{}, fmt.Errorf("failed to resolve symlinks: %w", err)
		}
		resolvedValue = resolved
	}

	rootWithSep := root.Value
	if !strings.HasSuffix(rootWithSep, string(filepath.Separator)) {
		rootWithSep = rootWithSep + string(filepath.Separator)
	}

	if resolvedValue != root.Value && !strings.HasPrefix(resolvedValue, rootWithSep) {
		return PathCfs{}, ErrResolvesOutsideRoot
	}

	relativePath := strings.TrimPrefix(resolvedValue, rootWithSep)
	if relativePath == resolvedValue {
		relativePath = strings.TrimPrefix(resolvedValue, root.Value)
		relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
	}

	cfsValue := filepath.ToSlash(relativePath)

	return PathCfs{Value: cfsValue}, nil
}
