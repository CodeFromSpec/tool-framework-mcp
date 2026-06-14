[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@mUOaDWmNceWf1XAxc8vRvD24Iqk)

# Package `pathutils`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils`

---

## Structs

```go
package pathutils

// PathCfs is a path in the Code from Spec standard format:
// forward slashes, relative to the project root, no ".." components,
// no drive letters, no leading slash, no backslashes.
type PathCfs struct {
	Value string
}

// PathOs is an absolute path in the operating system's native format.
// It is never exposed in the framework's public API.
type PathOs struct {
	Value string
}
```

---

## Error Sentinels

```go
package pathutils

import "errors"

var ErrCannotDetermineRoot    = errors.New("cannot determine project root")
var ErrPathEmpty              = errors.New("path is empty")
var ErrPathAbsolute           = errors.New("path must be relative (no leading slash or drive letter)")
var ErrPathContainsBackslash  = errors.New("path contains backslash characters")
var ErrDirectoryTraversal     = errors.New("path contains directory traversal components")
var ErrResolvesOutsideRoot    = errors.New("path resolves outside the project root")
```

---

## Functions

```go
package pathutils

// PathGetProjectRoot returns the project root as a PathOs,
// determined from the working directory of the process.
func PathGetProjectRoot() (*PathOs, error)

// PathValidateCfs validates that a value conforms to the PathCfs
// format rules. Returns an error describing the violation if not.
// Does not verify that the file exists or resolve symlinks.
func PathValidateCfs(value string) error

// PathCfsToOs validates a PathCfs and converts it to an absolute PathOs.
// The target file or directory does not need to exist.
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error)

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to
// the project root. The target file or directory does not need to exist.
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
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Project root:", root.Value)

	err = pathutils.PathValidateCfs("code-from-spec/SPEC/payments/fees/_node.md")
	if err != nil {
		log.Fatal(err)
	}

	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/SPEC/payments/fees/_node.md"}
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("OS path:", osPath.Value)

	backToCfs, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CFS path:", backToCfs.Value)
}
```
