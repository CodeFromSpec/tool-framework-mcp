[//]: # (code-from-spec: ROOT/golang/interfaces/os/path_utils@l4Cf-tdOGsA7MT4SfDQxVy2JD3Y)

# Package `pathutils`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
```

## Structs

```go
package pathutils

type PathCfs struct {
	Value string
}

type PathOs struct {
	Value string
}
```

## Error Sentinels

```go
package pathutils

import "errors"

var ErrCannotDetermineRoot = errors.New("cannot determine project root")
var ErrPathEmpty = errors.New("path is empty")
var ErrPathAbsolute = errors.New("path is absolute")
var ErrPathContainsBackslash = errors.New("path contains backslash")
var ErrDirectoryTraversal = errors.New("path contains directory traversal")
var ErrResolvesOutsideRoot = errors.New("path resolves outside project root")
```

## Functions

```go
package pathutils

// PathGetProjectRoot returns the project root as a PathOs,
// determined from the current working directory of the process.
func PathGetProjectRoot() (*PathOs, error)

// PathValidateCfs validates that a value conforms to the PathCfs format rules.
// It does not verify that the file exists or resolve symlinks.
func PathValidateCfs(value string) error

// PathCfsToOs validates a PathCfs and converts it to an absolute PathOs.
// The target file or directory does not need to exist.
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error)

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to
// the project root. The target file or directory does not need to exist.
func PathOsToCfs(osPath *PathOs) (*PathCfs, error)
```

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

	cfsPath := &pathutils.PathCfs{Value: "internal/filereader/filereader.go"}

	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("OS path:", osPath.Value)

	backCfs, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CFS path:", backCfs.Value)
}
```
