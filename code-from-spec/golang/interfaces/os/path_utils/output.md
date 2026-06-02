[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@gPMUYU_dQyxLRyyKFqK1kJ4Qi9A)

# Package `pathutils`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils`

## Types

```go
package pathutils

// PathCfs is a path in the Code from Spec standard format:
// forward-slash separator, relative to the project root,
// no ".." components, no drive letters, no leading "/",
// no backslashes.
type PathCfs struct {
	Value string
}

// PathOs is an absolute path in the operating system's native format.
// It is never exposed in the framework's public API.
type PathOs struct {
	Value string
}
```

## Error Sentinels

```go
package pathutils

import "errors"

var ErrCannotDetermineRoot    = errors.New("cannot determine project root")
var ErrPathEmpty              = errors.New("path is empty")
var ErrPathAbsolute           = errors.New("path must be relative")
var ErrPathContainsBackslash  = errors.New("path contains backslash")
var ErrDirectoryTraversal     = errors.New("path contains directory traversal")
var ErrResolvesOutsideRoot    = errors.New("path resolves outside project root")
```

## Functions

```go
package pathutils

// PathGetProjectRoot returns the project root as a PathOs,
// determined from the working directory of the process.
func PathGetProjectRoot() (*PathOs, error)

// PathValidateCfs validates that value conforms to the PathCfs
// format rules. Returns an error describing the first violation
// found. Does not verify that the file exists or resolve symlinks.
func PathValidateCfs(value string) error

// PathCfsToOs validates cfs_path and converts it to an absolute
// PathOs. The target does not need to exist on the filesystem.
func PathCfsToOs(cfs_path *PathCfs) (*PathOs, error)

// PathOsToCfs converts an absolute PathOs to a PathCfs relative
// to the project root. The target does not need to exist on the
// filesystem.
func PathOsToCfs(os_path *PathOs) (*PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	pathutils "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("project root:", root.Value)

	cfs := &pathutils.PathCfs{Value: "internal/filereader/filereader.go"}

	if err := pathutils.PathValidateCfs(cfs.Value); err != nil {
		log.Fatal(err)
	}

	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("absolute path:", osPath.Value)

	cfsRound, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("cfs path:", cfsRound.Value)
}
```
