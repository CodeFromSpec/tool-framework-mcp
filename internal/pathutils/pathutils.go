// code-from-spec: ROOT/golang/implementation/os/path_utils@96y-68Z4YL64ygTJTx-WtOxEtg4

package pathutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathCfs represents a path in the Code from Spec standard format.
// It uses forward slashes as separators, is always relative to the
// project root, and contains no ".." components, drive letters,
// leading slashes, or backslashes.
//
// Examples:
//   - "internal/filereader/filereader.go"
//   - "code-from-spec/functional/logic/os/file_reader/_node.md"
type PathCfs struct {
	Value string
}

// PathOs represents an absolute path in the operating system's native
// format, using the OS-specific separator. This type is never exposed
// in the framework's public API — it exists only inside the os/ layer
// for interacting with the filesystem.
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
	// cannot be read and the project root cannot be determined.
	ErrCannotDetermineRoot = errors.New("cannot determine root")

	// ErrPathEmpty is returned when the path value is empty.
	ErrPathEmpty = errors.New("path is empty")

	// ErrPathAbsolute is returned when the path starts with "/" or
	// a drive letter like "C:".
	ErrPathAbsolute = errors.New("path is absolute")

	// ErrPathContainsBackslash is returned when the path contains
	// "\" characters.
	ErrPathContainsBackslash = errors.New("path contains backslash")

	// ErrDirectoryTraversal is returned when the path contains ".."
	// components after normalization.
	ErrDirectoryTraversal = errors.New("directory traversal")

	// ErrResolvesOutsideRoot is returned when the resolved path falls
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

// PathValidateCfs validates that a value conforms to the PathCfs format
// rules. Returns an error describing the first violation found, if any.
// Follows OWASP guidance for path traversal prevention.
//
// Does not verify that the file exists or resolve symlinks — use
// PathCfsToOs for that.
//
// Possible errors:
//   - ErrPathEmpty
//   - ErrPathAbsolute
//   - ErrPathContainsBackslash
//   - ErrDirectoryTraversal
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

	// Normalize the path to resolve "." and ".." components.
	// filepath.Clean works on the native separator, so we use
	// path.Clean (which always uses "/") to normalize forward-slash paths.
	normalized := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))

	// After normalization, check for ".." components.
	for _, component := range strings.Split(normalized, "/") {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

// PathCfsToOs validates a PathCfs value and converts it to an absolute
// PathOs. This is the single entry point for going from framework paths
// to OS paths.
//
// The target file or directory does not need to exist. The conversion is
// purely path-based — it validates the format, converts separators, and
// checks containment, but does not require the path to resolve to an
// actual filesystem entry.
//
// Possible errors:
//   - ErrPathEmpty
//   - ErrPathAbsolute
//   - ErrPathContainsBackslash
//   - ErrDirectoryTraversal
//   - ErrResolvesOutsideRoot
//   - ErrCannotDetermineRoot
func PathCfsToOs(cfs_path *PathCfs) (*PathOs, error) {
	if err := PathValidateCfs(cfs_path.Value); err != nil {
		return nil, err
	}

	// Convert forward slashes to the OS-native separator.
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
			return nil, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}

		resolvedRoot, err := filepath.EvalSymlinks(root.Value)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
		}

		// Ensure the resolved path is contained within the resolved root.
		if !isContained(resolvedPath, resolvedRoot) {
			return nil, ErrResolvesOutsideRoot
		}
	}

	return &PathOs{Value: absPath}, nil
}

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the
// project root. Used internally by components that receive paths from
// the OS (e.g. directory listing).
//
// The target file or directory does not need to exist. The conversion is
// purely path-based.
//
// Possible errors:
//   - ErrResolvesOutsideRoot
//   - ErrCannotDetermineRoot
func PathOsToCfs(os_path *PathOs) (*PathCfs, error) {
	osPathValue := os_path.Value

	// If the path exists on disk, resolve symlinks.
	if _, statErr := os.Lstat(osPathValue); statErr == nil {
		resolved, err := filepath.EvalSymlinks(osPathValue)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}
		osPathValue = resolved
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	rootValue := root.Value

	// If the project root exists on disk, resolve its symlinks too.
	if _, statErr := os.Lstat(rootValue); statErr == nil {
		resolved, err := filepath.EvalSymlinks(rootValue)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
		}
		rootValue = resolved
	}

	// Verify containment.
	if !isContained(osPathValue, rootValue) {
		return nil, ErrResolvesOutsideRoot
	}

	// Remove the root prefix, including the trailing separator.
	rel := osPathValue[len(rootValue):]
	rel = strings.TrimPrefix(rel, string(filepath.Separator))

	// Convert OS separators to forward slashes.
	cfsValue := filepath.ToSlash(rel)

	return &PathCfs{Value: cfsValue}, nil
}

// isContained reports whether path is equal to base or is nested inside
// base. Both paths must already be cleaned/resolved before calling this.
func isContained(path, base string) bool {
	// Normalize separators for comparison.
	path = filepath.Clean(path)
	base = filepath.Clean(base)

	if path == base {
		return true
	}

	// Ensure the base ends with a separator so that a directory named
	// "/foo/bar" does not falsely match "/foo/barbaz".
	if !strings.HasSuffix(base, string(filepath.Separator)) {
		base += string(filepath.Separator)
	}

	return strings.HasPrefix(path, base)
}
