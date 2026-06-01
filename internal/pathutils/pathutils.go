// code-from-spec: ROOT/golang/implementation/os/path_utils@vLJM3Bi4uEvNV2aEVinoAHuSCIo

// Package pathutils provides path conversion and validation between the
// framework's canonical path format (CFS) and the operating system's
// native path format.
package pathutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathCfs represents a path in the Code from Spec standard format.
// It uses forward slash (/) as separator, is always relative to the
// project root, and contains no .. components, drive letters, leading
// slashes, or backslashes.
//
// Examples:
//   - "internal/filereader/filereader.go"
//   - "code-from-spec/functional/logic/os/file_reader/_node.md"
type PathCfs struct {
	Value string
}

// PathOs represents an absolute path in the operating system's native
// format. It uses the OS-specific separator and is always absolute.
//
// Examples (Unix):
//   - "/home/user/myproject/internal/filereader/filereader.go"
//
// Examples (Windows):
//   - `C:\Users\user\myproject\internal\filereader\filereader.go`
type PathOs struct {
	Value string
}

// ErrCannotDetermineRoot is returned when the working directory cannot
// be read and the project root cannot be determined.
var ErrCannotDetermineRoot = errors.New("cannot determine project root")

// ErrPathEmpty is returned when a CFS path value is empty.
var ErrPathEmpty = errors.New("path is empty")

// ErrPathAbsolute is returned when a CFS path starts with / or a
// drive letter (e.g. C:).
var ErrPathAbsolute = errors.New("path must be relative, not absolute")

// ErrPathContainsBackslash is returned when a CFS path contains
// backslash characters.
var ErrPathContainsBackslash = errors.New("path contains backslash characters")

// ErrDirectoryTraversal is returned when a CFS path contains ..
// components after normalization.
var ErrDirectoryTraversal = errors.New("path contains directory traversal components")

// ErrResolvesOutsideRoot is returned when a path resolves to a location
// outside the project root.
var ErrResolvesOutsideRoot = errors.New("path resolves outside the project root")

// PathGetProjectRoot returns the project root as a PathOs.
// The root is determined from the working directory of the process.
//
// Errors:
//   - ErrCannotDetermineRoot: the working directory cannot be read.
func PathGetProjectRoot() (*PathOs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
	}
	return &PathOs{Value: wd}, nil
}

// PathValidateCfs validates that a value conforms to the PathCfs
// format rules. Returns an error describing the violation if the
// value does not conform. Follows OWASP guidance for path traversal
// prevention.
//
// This function does not verify that the file exists or resolve
// symlinks. Use PathCfsToOs for that.
//
// Errors:
//   - ErrPathEmpty: the path value is empty.
//   - ErrPathAbsolute: the path starts with / or a drive letter like C:.
//   - ErrPathContainsBackslash: the path contains \ characters.
//   - ErrDirectoryTraversal: the path contains .. components after
//     normalization.
func PathValidateCfs(value string) error {
	if value == "" {
		return ErrPathEmpty
	}

	// Reject Unix-style absolute paths and Windows drive letters.
	if strings.HasPrefix(value, "/") || strings.Contains(value, ":") {
		return ErrPathAbsolute
	}

	// Reject backslashes.
	if strings.Contains(value, `\`) {
		return ErrPathContainsBackslash
	}

	// Normalize the path to resolve . and .. components.
	normalized := filepath.ToSlash(filepath.Clean(value))

	// Check each component for directory traversal.
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

// PathCfsToOs validates a PathCfs and converts it to an absolute PathOs.
// This is the single entry point for going from framework paths to OS
// paths. If validation fails, no conversion happens and an error is
// returned.
//
// The target file or directory does not need to exist. The conversion
// is purely path-based — it validates the format, converts separators,
// and checks containment, but does not require the path to resolve to
// an actual filesystem entry.
//
// Errors:
//   - ErrResolvesOutsideRoot: after resolving symlinks, the path is
//     outside the project root.
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

	// Convert forward slashes to OS-specific separators.
	nativePath := filepath.FromSlash(cfsPath.Value)

	// Join root with the converted relative path.
	absPath := filepath.Join(root.Value, nativePath)

	// Check whether the path exists on the filesystem.
	_, statErr := os.Stat(absPath)
	if statErr == nil {
		// Path exists — resolve symlinks and verify containment.
		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlinks for %q: %w", absPath, err)
		}

		rootWithSep := root.Value + string(filepath.Separator)
		if !strings.HasPrefix(realPath, rootWithSep) && realPath != root.Value {
			return nil, fmt.Errorf("%w: %q", ErrResolvesOutsideRoot, realPath)
		}
	}
	// If the path does not exist, skip the symlink check — containment
	// is guaranteed by the validation above (no .. allowed).

	return &PathOs{Value: absPath}, nil
}

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the
// project root. Used internally by components that receive paths from
// the OS (e.g. directory listing).
//
// The target file or directory does not need to exist. The conversion
// is purely path-based.
//
// Errors:
//   - ErrResolvesOutsideRoot: the path is not within the project root.
//   - (PathUtils.*): propagated from PathGetProjectRoot.
func PathOsToCfs(osPath *PathOs) (*PathCfs, error) {
	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	workingPath := osPath.Value

	// If the path exists, resolve symlinks.
	_, statErr := os.Stat(osPath.Value)
	if statErr == nil {
		realPath, err := filepath.EvalSymlinks(osPath.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlinks for %q: %w", osPath.Value, err)
		}
		workingPath = realPath
	}

	// Verify containment: working path must be within the project root.
	rootWithSep := root.Value + string(filepath.Separator)
	if workingPath != root.Value && !strings.HasPrefix(workingPath, rootWithSep) {
		return nil, fmt.Errorf("%w: %q", ErrResolvesOutsideRoot, workingPath)
	}

	// Compute the relative portion.
	var relativePath string
	if workingPath == root.Value {
		relativePath = ""
	} else {
		relativePath = strings.TrimPrefix(workingPath, rootWithSep)
	}

	// Convert OS separators to forward slashes.
	cfsValue := filepath.ToSlash(relativePath)

	return &PathCfs{Value: cfsValue}, nil
}
