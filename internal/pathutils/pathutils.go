// code-from-spec: ROOT/golang/implementation/os/path_utils@dVC1IZHgjZvB6VDrCRpvAMjGMBo

// Package pathutils provides path conversion and validation utilities for the
// Code from Spec framework. All framework-facing paths use the PathCfs format;
// OS-level operations use PathOs.
package pathutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathCfs is a path in the Code from Spec standard format:
//   - Forward slash (/) as separator, always.
//   - Relative to the project root.
//   - No .. components, no drive letters, no leading /, no backslashes.
//
// This is the only path format used in the framework's public API —
// in frontmatter fields (outputs, external, input), in logical names,
// and in tool parameters.
//
// Examples:
//   - internal/filereader/filereader.go
//   - code-from-spec/functional/logic/os/file_reader/_node.md
type PathCfs struct {
	Value string
}

// PathOs is an absolute path in the operating system's native format:
//   - OS-specific separator (/ on Unix, \ on Windows).
//   - Always absolute.
//
// This type is never exposed in the framework's public API.
// It exists only inside the os/ layer for interacting with the filesystem.
//
// Examples:
//   - /home/user/myproject/internal/filereader/filereader.go  (Unix)
//   - C:\Users\user\myproject\internal\filereader\filereader.go  (Windows)
type PathOs struct {
	Value string
}

// ErrCannotDetermineRoot is returned when the working directory cannot be read.
var ErrCannotDetermineRoot = errors.New("cannot determine project root")

// ErrPathEmpty is returned when a PathCfs value is empty.
var ErrPathEmpty = errors.New("path is empty")

// ErrPathAbsolute is returned when a PathCfs value starts with / or a drive letter like C:.
var ErrPathAbsolute = errors.New("path must be relative, not absolute")

// ErrPathContainsBackslash is returned when a PathCfs value contains \ characters.
var ErrPathContainsBackslash = errors.New("path contains backslash")

// ErrDirectoryTraversal is returned when a PathCfs value contains .. components
// after normalization.
var ErrDirectoryTraversal = errors.New("path contains directory traversal")

// ErrResolvesOutsideRoot is returned when a path resolves to a location outside
// the project root.
var ErrResolvesOutsideRoot = errors.New("path resolves outside project root")

// PathGetProjectRoot returns the project root as a PathOs.
// The root is determined from the working directory of the process.
//
// Returns ErrCannotDetermineRoot if the working directory cannot be read.
func PathGetProjectRoot() (*PathOs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
	}
	return &PathOs{Value: wd}, nil
}

// PathValidateCfs validates that a value conforms to the PathCfs format rules.
// Raises an error describing the first violation found.
// Follows OWASP guidance for path traversal prevention.
//
// Use this for sanity checks on parameters received from callers.
// Does not verify that the file exists or resolve symlinks — use PathCfsToOs for that.
//
// Errors:
//   - ErrPathEmpty: the path value is empty.
//   - ErrPathAbsolute: the path starts with / or a drive letter like C:.
//   - ErrPathContainsBackslash: the path contains \ characters.
//   - ErrDirectoryTraversal: the path contains .. components after normalization.
func PathValidateCfs(value string) error {
	if value == "" {
		return ErrPathEmpty
	}

	// Reject Unix-style absolute paths and Windows drive letters.
	if strings.HasPrefix(value, "/") || strings.Contains(value, ":") {
		return ErrPathAbsolute
	}

	// Reject backslashes — not allowed in CFS paths.
	if strings.Contains(value, `\`) {
		return ErrPathContainsBackslash
	}

	// Normalize to resolve . and .. components, then check for traversal.
	normalized := filepath.ToSlash(filepath.Clean(value))
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

// PathCfsToOs validates a PathCfs and converts it to an absolute PathOs.
// This is the single entry point for going from framework paths to OS paths.
// If validation fails, no conversion happens — an error is returned.
//
// The target file or directory does not need to exist.
// The conversion is purely path-based: it validates the format, converts
// separators, and checks containment, but does not require the path to
// resolve to an actual filesystem entry.
//
// Errors:
//   - ErrResolvesOutsideRoot: after resolving symlinks, the path is outside the project root.
//   - (PathUtils.*): propagated from PathValidateCfs.
//   - (PathUtils.*): propagated from PathGetProjectRoot.
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error) {
	if err := PathValidateCfs(cfsPath.Value); err != nil {
		return nil, err
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	// Convert forward slashes to the OS separator and join with root.
	nativePath := filepath.FromSlash(cfsPath.Value)
	absPath := filepath.Join(root.Value, nativePath)

	// If the path exists on disk, resolve symlinks and verify containment.
	if _, statErr := os.Lstat(absPath); statErr == nil {
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks: %w", err)
		}
		if !strings.HasPrefix(resolved, root.Value) {
			return nil, fmt.Errorf("%w: %s", ErrResolvesOutsideRoot, resolved)
		}
	}

	return &PathOs{Value: absPath}, nil
}

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the project root.
// Used internally by components that receive paths from the OS (e.g. directory listing).
//
// The target file or directory does not need to exist.
// The conversion is purely path-based.
//
// Errors:
//   - ErrResolvesOutsideRoot: the path is not within the project root.
//   - (PathUtils.*): propagated from PathGetProjectRoot.
func PathOsToCfs(osPath *PathOs) (*PathCfs, error) {
	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	workingValue := osPath.Value

	// If the path exists on disk, resolve symlinks first.
	if _, statErr := os.Lstat(workingValue); statErr == nil {
		resolved, err := filepath.EvalSymlinks(workingValue)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks: %w", err)
		}
		workingValue = resolved
	}

	// Verify the path is within the project root.
	rootWithSep := root.Value + string(filepath.Separator)
	if workingValue != root.Value && !strings.HasPrefix(workingValue, rootWithSep) {
		return nil, fmt.Errorf("%w: %s", ErrResolvesOutsideRoot, workingValue)
	}

	// Strip the root prefix (and trailing separator) to get the relative path.
	relative := strings.TrimPrefix(workingValue, root.Value)
	relative = strings.TrimPrefix(relative, string(filepath.Separator))

	// Convert OS separators to forward slashes for CFS format.
	cfsValue := filepath.ToSlash(relative)

	return &PathCfs{Value: cfsValue}, nil
}
