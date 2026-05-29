[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@NhE_QwS2ITwuPCup-RV8tKWOTSQ)

# Interface: `pathutils`

**Package:** `package pathutils`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"`

---

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
```

---

## Error Sentinels

```go
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
```

---

## Functions

```go
// PathGetProjectRoot returns the project root as a PathOs, determined
// from the working directory of the process.
//
// Returns ErrCannotDetermineRoot if the working directory cannot be read.
func PathGetProjectRoot() (*PathOs, error)

// PathValidateCfs validates that a string value conforms to the PathCfs
// format rules. It does not verify that the file exists or resolve
// symlinks. Follows OWASP guidance for path traversal prevention.
//
// Returns one of the following errors if validation fails:
//   - ErrPathIsEmpty: the value is empty.
//   - ErrPathIsAbsolute: the value starts with "/" or a drive letter.
//   - ErrPathContainsBackslash: the value contains "\" characters.
//   - ErrDirectoryTraversal: the value contains ".." after normalization.
func PathValidateCfs(value string) error

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
func PathCfsToOs(cfs_path *PathCfs) (*PathOs, error)

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
func PathOsToCfs(os_path *PathOs) (*PathCfs, error)
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
    // Retrieve the project root.
    root, err := pathutils.PathGetProjectRoot()
    if err != nil {
        log.Fatalf("could not get project root: %v", err)
    }
    fmt.Println("Project root:", root.Value)

    // Validate a CFS path before using it.
    rawPath := "internal/filereader/filereader.go"
    if err := pathutils.PathValidateCfs(rawPath); err != nil {
        log.Fatalf("invalid CFS path: %v", err)
    }

    // Convert a CFS path to an OS path.
    cfsPath := &pathutils.PathCfs{Value: rawPath}
    osPath, err := pathutils.PathCfsToOs(cfsPath)
    if err != nil {
        log.Fatalf("could not convert CFS path to OS path: %v", err)
    }
    fmt.Println("OS path:", osPath.Value)

    // Convert an OS path back to a CFS path.
    cfsBack, err := pathutils.PathOsToCfs(osPath)
    if err != nil {
        log.Fatalf("could not convert OS path to CFS path: %v", err)
    }
    fmt.Println("CFS path:", cfsBack.Value)
}
```
