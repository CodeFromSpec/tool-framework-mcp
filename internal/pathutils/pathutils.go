// code-from-spec: ROOT/golang/implementation/os/path_utils@Xo9FYx4jreFhNUB1u18fNlg1k9s

package pathutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathCfs represents a path in the Code from Spec standard format.
// It uses forward slashes as separators, is relative to the project
// root, and contains no ".." components, drive letters, leading
// slashes, or backslashes.
type PathCfs struct {
	Value string
}

// PathOs represents an absolute path in the operating system's native
// format. It uses OS-specific separators and is always absolute.
// This type is never exposed in the framework's public API.
type PathOs struct {
	Value string
}

var (
	// ErrCannotDetermineRoot is returned when the working directory
	// cannot be read.
	ErrCannotDetermineRoot = errors.New("cannot determine root")

	// ErrPathEmpty is returned when a CFS path value is empty.
	ErrPathEmpty = errors.New("path is empty")

	// ErrPathAbsolute is returned when a CFS path starts with "/"
	// or a drive letter like "C:".
	ErrPathAbsolute = errors.New("path is absolute")

	// ErrPathContainsBackslash is returned when a CFS path contains
	// backslash characters.
	ErrPathContainsBackslash = errors.New("path contains backslash")

	// ErrDirectoryTraversal is returned when a CFS path contains ".."
	// components after normalization.
	ErrDirectoryTraversal = errors.New("directory traversal")

	// ErrResolvesOutsideRoot is returned when a path resolves to a
	// location outside the project root.
	ErrResolvesOutsideRoot = errors.New("resolves outside root")
)

// PathGetProjectRoot returns the project root as a PathOs, determined
// from the working directory of the process.
//
// Returns ErrCannotDetermineRoot if the working directory cannot be read.
func PathGetProjectRoot() (*PathOs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
	}
	return &PathOs{Value: wd}, nil
}

// PathValidateCfs validates that a value conforms to the PathCfs format
// rules. Returns an error describing the first violation found, if any.
// Follows OWASP guidance for path traversal prevention.
//
// This function does not verify that the file exists or resolve symlinks.
// Use PathCfsToOs for that.
//
// Possible errors: ErrPathEmpty, ErrPathAbsolute, ErrPathContainsBackslash,
// ErrDirectoryTraversal.
func PathValidateCfs(value string) error {
	if value == "" {
		return ErrPathEmpty
	}

	if strings.HasPrefix(value, "/") || strings.Contains(value, ":") {
		return ErrPathAbsolute
	}

	if strings.Contains(value, `\`) {
		return ErrPathContainsBackslash
	}

	// Normalize to resolve "." and ".." components.
	normalized := filepath.ToSlash(filepath.Clean(value))

	// After normalization, check for ".." components.
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

// PathCfsToOs validates a PathCfs and converts it to an absolute PathOs.
// This is the single entry point for going from framework paths to OS paths.
// If validation fails, no conversion happens and an error is returned.
//
// The target file or directory does not need to exist. The conversion is
// purely path-based: it validates the format, converts separators, and
// checks containment, but does not require the path to resolve to an
// actual filesystem entry.
//
// Possible errors: ErrPathEmpty, ErrPathAbsolute, ErrPathContainsBackslash,
// ErrDirectoryTraversal, ErrResolvesOutsideRoot, ErrCannotDetermineRoot.
func PathCfsToOs(cfs_path *PathCfs) (*PathOs, error) {
	if err := PathValidateCfs(cfs_path.Value); err != nil {
		return nil, err
	}

	// Replace forward slashes with the OS-native separator.
	nativePath := filepath.FromSlash(cfs_path.Value)

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	absPath := filepath.Join(root.Value, nativePath)

	// If the path exists on disk, resolve symlinks and verify containment.
	if _, statErr := os.Lstat(absPath); statErr == nil {
		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks in path: %w", err)
		}

		resolvedRoot, err := filepath.EvalSymlinks(root.Value)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks in root: %w", err)
		}

		// Ensure resolvedRoot ends with a separator for prefix matching.
		rootPrefix := resolvedRoot + string(filepath.Separator)
		if resolvedPath != resolvedRoot && !strings.HasPrefix(resolvedPath, rootPrefix) {
			return nil, ErrResolvesOutsideRoot
		}
	}

	return &PathOs{Value: absPath}, nil
}

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the
// project root. Used internally by components that receive paths from the
// OS (e.g. directory listing).
//
// The target file or directory does not need to exist. The conversion is
// purely path-based.
//
// Possible errors: ErrResolvesOutsideRoot, ErrCannotDetermineRoot.
func PathOsToCfs(os_path *PathOs) (*PathCfs, error) {
	resolvedPath := os_path.Value

	// If path exists on disk, resolve symlinks.
	if _, statErr := os.Lstat(resolvedPath); statErr == nil {
		var err error
		resolvedPath, err = filepath.EvalSymlinks(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks in path: %w", err)
		}
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	resolvedRoot := root.Value

	// If root exists on disk, resolve symlinks.
	if _, statErr := os.Lstat(resolvedRoot); statErr == nil {
		resolvedRoot, err = filepath.EvalSymlinks(resolvedRoot)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks in root: %w", err)
		}
	}

	// Verify the path is within the root.
	rootPrefix := resolvedRoot + string(filepath.Separator)
	if resolvedPath == resolvedRoot {
		return &PathCfs{Value: ""}, nil
	}
	if !strings.HasPrefix(resolvedPath, rootPrefix) {
		return nil, ErrResolvesOutsideRoot
	}

	// Remove the root prefix to get the relative path.
	relativePath := strings.TrimPrefix(resolvedPath, rootPrefix)

	// Convert OS separators to forward slashes.
	cfsValue := filepath.ToSlash(relativePath)

	return &PathCfs{Value: cfsValue}, nil
}
