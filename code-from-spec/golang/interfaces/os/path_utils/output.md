[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@ZJe3r0ztGKe39vfe1vI9lwsdF0Q)

# Package `pathutils`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
```

Provides path conversion and validation utilities for the Code from Spec framework. All framework-facing paths use the `PathCfs` format; OS-level operations use `PathOs`.

---

## Structs

```go
package pathutils

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
```

---

## Error Sentinels

```go
package pathutils

import "errors"

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
```

---

## Functions

```go
package pathutils

// PathGetProjectRoot returns the project root as a PathOs.
// The root is determined from the working directory of the process.
//
// Returns ErrCannotDetermineRoot if the working directory cannot be read.
func PathGetProjectRoot() (*PathOs, error)

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
func PathValidateCfs(value string) error

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
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error)

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the project root.
// Used internally by components that receive paths from the OS (e.g. directory listing).
//
// The target file or directory does not need to exist.
// The conversion is purely path-based.
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
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	// Validate a CFS-format path before using it.
	if err := pathutils.PathValidateCfs("internal/filereader/filereader.go"); err != nil {
		log.Fatalf("invalid path: %v", err)
	}

	// Convert a CFS path to an OS-native absolute path.
	cfs := &pathutils.PathCfs{Value: "internal/filereader/filereader.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		log.Fatalf("PathCfsToOs: %v", err)
	}
	fmt.Println("OS path:", osPath.Value)

	// Convert an OS-native absolute path back to a CFS path.
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatalf("PathGetProjectRoot: %v", err)
	}
	fmt.Println("Project root:", root.Value)

	reconstructed := &pathutils.PathOs{Value: osPath.Value}
	cfsBack, err := pathutils.PathOsToCfs(reconstructed)
	if err != nil {
		log.Fatalf("PathOsToCfs: %v", err)
	}
	fmt.Println("CFS path:", cfsBack.Value)
}
```
