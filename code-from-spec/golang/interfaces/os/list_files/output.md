[//]: # (code-from-spec: ROOT/golang/interfaces/os/list_files@Zxs88k51iT5lI6_P5QhqS3w_xM4)

# Package `listfiles`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
```

Provides recursive directory listing, returning all files under a given directory as `PathCfs` values sorted alphabetically.

---

## Error Sentinels

```go
package listfiles

import "errors"

// ErrDirectoryNotFound is returned when the given directory does not exist.
var ErrDirectoryNotFound = errors.New("directory not found")

// ErrWalkError is returned when a filesystem error occurs while traversing
// the directory tree.
var ErrWalkError = errors.New("walk error")
```

---

## Functions

```go
package listfiles

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ListFiles returns all files (not directories) found recursively under the
// given directory. Results are PathCfs values sorted alphabetically. If the
// directory exists but contains no files, returns an empty slice.
//
// Errors:
//   - ErrDirectoryNotFound: the directory does not exist.
//   - ErrWalkError: a filesystem error occurred while traversing.
//   - (PathUtils.*): propagated from PathCfsToOs.
//   - (PathUtils.*): propagated from PathOsToCfs.
func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	dir := &pathutils.PathCfs{Value: "internal/listfiles"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		log.Fatalf("ListFiles: %v", err)
	}

	for _, f := range files {
		fmt.Println(f.Value)
	}
}
```
