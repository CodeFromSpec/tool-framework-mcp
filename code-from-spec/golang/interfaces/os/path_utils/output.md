[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@aSkWddeikRDcMdF6HMrgF6e50CI)

# Package `pathutils`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
```

Package `pathutils` provides path conversion and validation between the framework's canonical path format (CFS) and the operating system's native path format.

---

## Structs

```go
package pathutils

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
```

---

## Error Sentinels

```go
package pathutils

import "errors"

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
```

---

## Functions

```go
package pathutils

// PathGetProjectRoot returns the project root as a PathOs.
// The root is determined from the working directory of the process.
//
// Errors:
//   - ErrCannotDetermineRoot: the working directory cannot be read.
func PathGetProjectRoot() (*PathOs, error)

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
func PathValidateCfs(value string) error

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
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error)

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
func PathOsToCfs(osPath *PathOs) (*PathCfs, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	// Get the project root.
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatalf("failed to determine project root: %v", err)
	}
	fmt.Println("project root:", root.Value)

	// Validate a CFS path before using it.
	rawPath := "internal/filereader/filereader.go"
	if err := pathutils.PathValidateCfs(rawPath); err != nil {
		if errors.Is(err, pathutils.ErrPathEmpty) {
			log.Fatal("path must not be empty")
		}
		if errors.Is(err, pathutils.ErrDirectoryTraversal) {
			log.Fatal("path traversal detected")
		}
		log.Fatalf("invalid path: %v", err)
	}

	// Convert a CFS path to an OS path.
	cfsPath := &pathutils.PathCfs{Value: rawPath}
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		if errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			log.Fatal("path escapes project root")
		}
		log.Fatalf("conversion failed: %v", err)
	}
	fmt.Println("os path:", osPath.Value)

	// Convert an OS path back to a CFS path.
	backToCfs, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		if errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			log.Fatal("os path is outside project root")
		}
		log.Fatalf("reverse conversion failed: %v", err)
	}
	fmt.Println("cfs path:", backToCfs.Value)
}
```
