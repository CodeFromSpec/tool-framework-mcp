// code-from-spec: ROOT/golang/implementation/os/path_utils@TNSwpe_gorGjksRfvyqmPEU5IX8

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

// PathOs represents an absolute path in the operating system's native
// format. It uses the OS-specific separator and is always absolute.
// This type is used only within the os/ layer for filesystem interaction.
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
	// cannot be read.
	ErrCannotDetermineRoot = errors.New("cannot determine root")

	// ErrPathEmpty is returned when a PathCfs value is an empty string.
	ErrPathEmpty = errors.New("path is empty")

	// ErrPathAbsolute is returned when a PathCfs value starts with "/"
	// or a drive letter such as "C:".
	ErrPathAbsolute = errors.New("path is absolute")

	// ErrPathContainsBackslash is returned when a PathCfs value contains
	// backslash characters.
	ErrPathContainsBackslash = errors.New("path contains backslash")

	// ErrDirectoryTraversal is returned when a PathCfs value contains
	// ".." components after normalization.
	ErrDirectoryTraversal = errors.New("directory traversal")

	// ErrResolvesOutsideRoot is returned when a path, after resolution,
	// falls outside the project root.
	ErrResolvesOutsideRoot = errors.New("resolves outside root")
)

// PathGetProjectRoot returns the project root as a PathOs, determined
// from the working directory of the process.
//
// Returns ErrCannotDetermineRoot if the working directory cannot be read.
func PathGetProjectRoot() (PathOs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return PathOs{}, fmt.Errorf("%w: %w", ErrCannotDetermineRoot, err)
	}
	return PathOs{Value: wd}, nil
}

// PathValidateCfs validates that value conforms to the PathCfs format
// rules. Returns an error describing the first violation found, if any.
// Follows OWASP guidance for path traversal prevention.
//
// This function does not verify that the file exists or resolve symlinks.
// Use PathCfsToOs for that.
//
// Possible errors:
//   - ErrPathEmpty: the value is an empty string.
//   - ErrPathAbsolute: the value starts with "/" or a drive letter like "C:".
//   - ErrPathContainsBackslash: the value contains "\" characters.
//   - ErrDirectoryTraversal: the value contains ".." components after
//     normalization.
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

	// Normalize the path by resolving "." and ".." components.
	// Use ToSlash/FromSlash to ensure filepath.Clean works correctly,
	// then convert back for component inspection.
	normalized := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))

	// Check each component for ".."
	components := strings.Split(normalized, "/")
	for _, component := range components {
		if component == ".." {
			return ErrDirectoryTraversal
		}
	}

	return nil
}

// PathCfsToOs validates cfs_path and converts it to an absolute PathOs.
// This is the single entry point for converting framework paths to OS paths.
// If validation fails, no conversion is performed and an error is returned.
//
// The target file or directory does not need to exist. The conversion is
// purely path-based: it validates the format, converts separators, and
// checks containment, but does not require the path to resolve to an actual
// filesystem entry.
//
// Possible errors:
//   - ErrPathEmpty, ErrPathAbsolute, ErrPathContainsBackslash,
//     ErrDirectoryTraversal: propagated from PathValidateCfs.
//   - ErrResolvesOutsideRoot: after resolving symlinks, the path is outside
//     the project root.
func PathCfsToOs(cfs_path PathCfs) (PathOs, error) {
	if err := PathValidateCfs(cfs_path.Value); err != nil {
		return PathOs{}, err
	}

	// Replace forward slashes with the OS-native separator.
	nativePath := filepath.FromSlash(cfs_path.Value)

	root, err := PathGetProjectRoot()
	if err != nil {
		return PathOs{}, err
	}

	absPath := filepath.Join(root.Value, nativePath)

	// Only do symlink resolution and containment check if the path exists.
	_, statErr := os.Lstat(absPath)
	if statErr == nil {
		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return PathOs{}, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}

		resolvedRoot, err := filepath.EvalSymlinks(root.Value)
		if err != nil {
			return PathOs{}, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}

		// Ensure the resolved path starts with the resolved root.
		// Add a separator to the root to avoid partial directory name matches.
		rootWithSep := resolvedRoot + string(filepath.Separator)
		if resolvedPath != resolvedRoot && !strings.HasPrefix(resolvedPath, rootWithSep) {
			return PathOs{}, ErrResolvesOutsideRoot
		}

		return PathOs{Value: resolvedPath}, nil
	}

	return PathOs{Value: absPath}, nil
}

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the
// project root. Used internally by components that receive paths from the OS
// (e.g. directory listing).
//
// The target file or directory does not need to exist. The conversion is
// purely path-based.
//
// Possible errors:
//   - ErrResolvesOutsideRoot: the path is not within the project root.
func PathOsToCfs(os_path PathOs) (PathCfs, error) {
	resolvedPath := os_path.Value

	// If the path exists, resolve symlinks.
	_, statErr := os.Lstat(resolvedPath)
	if statErr == nil {
		rp, err := filepath.EvalSymlinks(resolvedPath)
		if err != nil {
			return PathCfs{}, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}
		resolvedPath = rp
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		return PathCfs{}, err
	}

	resolvedRoot := root.Value

	// If the project root exists, resolve its symlinks too.
	_, rootStatErr := os.Lstat(resolvedRoot)
	if rootStatErr == nil {
		rr, err := filepath.EvalSymlinks(resolvedRoot)
		if err != nil {
			return PathCfs{}, fmt.Errorf("%w: %w", ErrResolvesOutsideRoot, err)
		}
		resolvedRoot = rr
	}

	// Check containment.
	rootWithSep := resolvedRoot + string(filepath.Separator)
	if resolvedPath == resolvedRoot {
		// The path is the root itself — relative path is empty, return ".".
		return PathCfs{Value: "."}, nil
	}
	if !strings.HasPrefix(resolvedPath, rootWithSep) {
		return PathCfs{}, ErrResolvesOutsideRoot
	}

	// Strip the root prefix and the following separator.
	relative := resolvedPath[len(rootWithSep):]

	// Convert OS-native separators to forward slashes.
	cfspath := filepath.ToSlash(relative)

	return PathCfs{Value: cfspath}, nil
}
