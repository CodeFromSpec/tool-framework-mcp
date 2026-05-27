[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@2LF9diqBucmb88jRxh5J7gS96rE)

# Interface: `pathutils`

## Package

```go
package pathutils
```

## Structs

```go
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
```

## Error Sentinels

```go
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
```

## Functions

```go
// PathGetProjectRoot returns the project root as a PathOs, determined
// from the working directory of the process.
//
// Returns ErrCannotDetermineRoot if the working directory cannot be read.
func PathGetProjectRoot() (PathOs, error)

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
func PathValidateCfs(value string) error

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
func PathCfsToOs(cfs_path PathCfs) (PathOs, error)

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the
// project root. Used internally by components that receive paths from the OS
// (e.g. directory listing).
//
// The target file or directory does not need to exist. The conversion is
// purely path-based.
//
// Possible errors:
//   - ErrResolvesOutsideRoot: the path is not within the project root.
func PathOsToCfs(os_path PathOs) (PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/os/pathutils"
)

func main() {
	// Get the project root from the process working directory.
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}
	fmt.Println("project root:", root.Value)

	// Validate a CFS path without converting it.
	if err := pathutils.PathValidateCfs("internal/filereader/filereader.go"); err != nil {
		log.Fatalf("invalid cfs path: %v", err)
	}

	// Convert a CFS path to an OS-native absolute path.
	cfs := pathutils.PathCfs{Value: "internal/filereader/filereader.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		log.Fatalf("failed to convert cfs to os path: %v", err)
	}
	fmt.Println("os path:", osPath.Value)

	// Convert an OS path back to a CFS path.
	cfsBack, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		log.Fatalf("failed to convert os path to cfs: %v", err)
	}
	fmt.Println("cfs path:", cfsBack.Value)
}
```
