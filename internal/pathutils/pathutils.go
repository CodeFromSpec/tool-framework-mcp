// code-from-spec: ROOT/golang/implementation/os/path_utils@V4-WCYptZW2vZH4wVS9WN9Dp_y0

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
	if strings.HasPrefix(value, "/") || (len(value) >= 2 && value[1] == ':') {
		return ErrPathAbsolute
	}

	if strings.Contains(value, "\\") {
		return ErrPathContainsBackslash
	}

	// Normalize to resolve "." and ".." components.
	normalized := filepath.ToSlash(filepath.Clean(value))

	// After normalization, check each component for "..".
	components := strings.Split(normalized, "/")
	for _, component := range components {
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

	joined := filepath.Join(root.Value, nativePath)

	// If the joined path exists, resolve symlinks and verify containment.
	if _, statErr := os.Stat(joined); statErr == nil {
		resolved, err := filepath.EvalSymlinks(joined)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}
		if !isContainedIn(resolved, root.Value) {
			return nil, ErrResolvesOutsideRoot
		}
	}

	return &PathOs{Value: joined}, nil
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
	root, err := PathGetProjectRoot()
	if err != nil {
		return nil, err
	}

	targetPath := os_path.Value

	// If the path exists on disk, resolve symlinks.
	if _, statErr := os.Stat(targetPath); statErr == nil {
		resolved, err := filepath.EvalSymlinks(targetPath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}
		targetPath = resolved
	}

	if !isContainedIn(targetPath, root.Value) {
		return nil, ErrResolvesOutsideRoot
	}

	// Remove the root prefix and any trailing separator that follows it.
	relative := targetPath[len(root.Value):]
	relative = strings.TrimPrefix(relative, string(filepath.Separator))

	// Convert OS-native separators to forward slashes.
	cfsValue := filepath.ToSlash(relative)

	return &PathCfs{Value: cfsValue}, nil
}

// isContainedIn reports whether path is contained within base,
// meaning path equals base or path starts with base followed by
// the OS path separator.
func isContainedIn(path, base string) bool {
	if path == base {
		return true
	}
	return strings.HasPrefix(path, base+string(filepath.Separator))
}
