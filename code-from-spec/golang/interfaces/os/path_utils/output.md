[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@ALSUij1PXdV5NuNOlsnqqUv0iUg)

# Interface: `pathutils`

## Package

```go
package pathutils
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
```

---

## Struct Definitions

```go
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
```

---

## Error Sentinels

```go
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
```

---

## Functions

```go
// PathGetProjectRoot returns the project root as a PathOs, determined
// from the working directory of the process.
//
// Returns ErrCannotDetermineRoot if the working directory cannot be read.
func PathGetProjectRoot() (*PathOs, error)

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
func PathValidateCfs(value string) error

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
func PathCfsToOs(cfs_path *PathCfs) (*PathOs, error)

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
func PathOsToCfs(os_path *PathOs) (*PathCfs, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

func main() {
	// Get the project root.
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Project root:", root.Value)

	// Validate a CFS path before use.
	if err := pathutils.PathValidateCfs("internal/filereader/filereader.go"); err != nil {
		log.Fatal(err)
	}

	// Convert a CFS path to an OS path.
	cfs := &pathutils.PathCfs{Value: "internal/filereader/filereader.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("OS path:", osPath.Value)

	// Convert an OS path back to a CFS path.
	cfsBack, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CFS path:", cfsBack.Value)
}
```
