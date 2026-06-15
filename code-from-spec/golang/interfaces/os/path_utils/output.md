[//]: # (code-from-spec: SPEC/golang/interfaces/os/path_utils@b_xbAfj8SOcHW7JKI-SWlNvdyT4)

# Package `pathutils`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils`

## Types

```go
package pathutils

// PathCfs is a path in the Code from Spec standard format:
// forward-slash separated, relative to the project root,
// no ".." components, no drive letters, no leading "/", no backslashes.
type PathCfs struct {
	Value string
}

// PathOs is an absolute path in the operating system's native format.
// This type is never exposed in the framework's public API.
type PathOs struct {
	Value string
}
```

## Error Sentinels

```go
package pathutils

import "errors"

var ErrCannotDetermineRoot   = errors.New("cannot determine project root")
var ErrPathEmpty             = errors.New("path is empty")
var ErrPathAbsolute          = errors.New("path must not be absolute")
var ErrPathContainsBackslash = errors.New("path must not contain backslashes")
var ErrDirectoryTraversal    = errors.New("path contains directory traversal components")
var ErrResolvesOutsideRoot   = errors.New("path resolves outside the project root")
```

## Functions

```go
package pathutils

// PathGetProjectRoot returns the project root as a PathOs,
// determined from the working directory of the process.
func PathGetProjectRoot() (*PathOs, error)

// PathValidateCfs validates that value conforms to the PathCfs format rules.
// Returns an error describing the violation if the value is not valid.
// Does not verify that the file exists or resolve symlinks.
func PathValidateCfs(value string) error

// PathCfsToOs validates cfs_path and converts it to an absolute PathOs.
// This is the single entry point for going from framework paths to OS paths.
// The target file or directory does not need to exist.
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error)

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the
// project root. Used internally by components that receive paths from the OS.
// The target file or directory does not need to exist.
func PathOsToCfs(osPath *PathOs) (*PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Project root:", root.Value)

	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/_node.md"}

	if err := pathutils.PathValidateCfs(cfsPath.Value); err != nil {
		log.Fatal(err)
	}

	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("OS path:", osPath.Value)

	roundTripped, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CFS path:", roundTripped.Value)
}
```
