// code-from-spec: ROOT/golang/implementation/os/path_utils@ufGrf2QahZC6yHfmweJ_Z3Spwy4

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
//
// Examples:
//   - "internal/filereader/filereader.go"
//   - "code-from-spec/functional/logic/os/file_reader/_node.md"
type PathCfs struct {
	Value string
}

// PathOs represents an absolute path in the operating system's
// native format. It uses the OS-specific separator and is always
// absolute.
//
// Examples (Unix):
//   - "/home/user/myproject/internal/filereader/filereader.go"
//
// Examples (Windows):
//   - `C:\Users\user\myproject\internal\filereader\filereader.go`
type PathOs struct {
	Value string
}

var (
	// ErrCannotDetermineRoot is returned when the working directory
	// cannot be read to determine the project root.
	ErrCannotDetermineRoot = errors.New("cannot determine root")

	// ErrPathIsEmpty is returned when a CFS path value is empty.
	ErrPathIsEmpty = errors.New("path is empty")

	// ErrPathIsAbsolute is returned when a CFS path starts with "/"
	// or a drive letter (e.g. "C:").
	ErrPathIsAbsolute = errors.New("path is absolute")

	// ErrPathContainsBackslash is returned when a CFS path contains
	// backslash characters.
	ErrPathContainsBackslash = errors.New("path contains backslash")

	// ErrDirectoryTraversal is returned when a CFS path contains ".."
	// components after normalization.
	ErrDirectoryTraversal = errors.New("directory traversal")

	// ErrResolvesOutsideRoot is returned when a resolved path falls
	// outside the project root.
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

// PathValidateCfs validates that a string value conforms to the PathCfs
// format rules. It does not verify that the file exists or resolve
// symlinks. Follows OWASP guidance for path traversal prevention.
//
// Returns one of the following errors if validation fails:
//   - ErrPathIsEmpty: the value is empty.
//   - ErrPathIsAbsolute: the value starts with "/" or a drive letter.
//   - ErrPathContainsBackslash: the value contains "\" characters.
//   - ErrDirectoryTraversal: the value contains ".." after normalization.
func PathValidateCfs(value string) error {
	if value == "" {
		return ErrPathIsEmpty
	}

	if strings.HasPrefix(value, "/") || strings.Contains(value, ":") {
		return ErrPathIsAbsolute
	}

	if strings.Contains(value, `\`) {
		return ErrPathContainsBackslash
	}

	// Normalize by cleaning the path (resolves "." and ".." components).
	normalized := filepath.ToSlash(filepath.Clean(value))

	// After normalization, check each component for "..".
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

// PathCfsToOs validates a PathCfs and converts it to an absolute PathOs.
// This is the single entry point for going from framework paths to OS paths.
//
// The target file or directory does not need to exist. The conversion is
// purely path-based: it validates format, converts separators, and checks
// containment, but does not require the path to resolve to an actual
// filesystem entry.
//
// Returns an error if:
//   - validation fails (errors from PathValidateCfs are propagated).
//   - the project root cannot be determined (ErrCannotDetermineRoot).
//   - after resolving symlinks, the path is outside the project root
//     (ErrResolvesOutsideRoot).
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error) {
	if err := PathValidateCfs(cfsPath.Value); err != nil {
		return nil, err
	}

	// Convert forward slashes to the OS path separator.
	nativePath := filepath.FromSlash(cfsPath.Value)

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	absPath := filepath.Join(root.Value, nativePath)

	// If the path exists on disk, resolve symlinks and verify containment.
	if _, err := os.Lstat(absPath); err == nil {
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks: %w", err)
		}

		// Ensure the resolved path starts with the project root.
		rootWithSep := ensureTrailingSeparator(root.Value)
		if resolved != root.Value && !strings.HasPrefix(resolved, rootWithSep) {
			return nil, fmt.Errorf("%w: %s", ErrResolvesOutsideRoot, resolved)
		}

		return &PathOs{Value: resolved}, nil
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
// Returns an error if:
//   - the project root cannot be determined (ErrCannotDetermineRoot).
//   - the path is not within the project root (ErrResolvesOutsideRoot).
func PathOsToCfs(osPath *PathOs) (*PathCfs, error) {
	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	candidate := osPath.Value

	// If the path exists on disk, resolve symlinks before checking containment.
	if _, err := os.Lstat(candidate); err == nil {
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return nil, fmt.Errorf("resolving symlinks: %w", err)
		}
		candidate = resolved
	}

	// Verify that the path is within the project root.
	rootWithSep := ensureTrailingSeparator(root.Value)
	if candidate != root.Value && !strings.HasPrefix(candidate, rootWithSep) {
		return nil, fmt.Errorf("%w: %s", ErrResolvesOutsideRoot, candidate)
	}

	// Compute the relative path by stripping the root prefix.
	rel := strings.TrimPrefix(candidate, root.Value)
	rel = strings.TrimPrefix(rel, string(filepath.Separator))

	// Convert OS separators to forward slashes.
	cfsValue := filepath.ToSlash(rel)

	return &PathCfs{Value: cfsValue}, nil
}

// ensureTrailingSeparator appends a filepath separator to the given path
// if it does not already end with one. This is used for reliable prefix
// checking to avoid false matches (e.g., "/foo/bar" matching "/foo/barbaz").
func ensureTrailingSeparator(path string) string {
	if strings.HasSuffix(path, string(filepath.Separator)) {
		return path
	}
	return path + string(filepath.Separator)
}
