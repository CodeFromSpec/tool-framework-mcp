[//]: # (code-from-spec: ROOT/golang/interfaces/os/list_files@fu18yPTnlODZM7TjbkGB_enHxzE)

# Package `listfiles`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
```

Package `listfiles` provides a function to recursively list all files under a
given directory, returning paths in the framework's canonical CFS format.

---

## Error Sentinels

```go
package listfiles

import "errors"

// ErrDirectoryNotFound is returned when the specified directory does not exist.
var ErrDirectoryNotFound = errors.New("directory not found")

// ErrWalkError is returned when a filesystem error occurs while traversing
// the directory tree.
var ErrWalkError = errors.New("filesystem walk error")
```

---

## Functions

```go
package listfiles

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ListFiles returns all files (not directories) found recursively under the
// directory identified by cfsPath. Results are returned as PathCfs values
// sorted alphabetically. If the directory exists but contains no files, an
// empty slice is returned.
//
// Errors:
//   - ErrDirectoryNotFound: the directory does not exist.
//   - ErrWalkError: a filesystem error occurred while traversing the tree.
//   - (PathUtils.*): propagated from pathutils.PathCfsToOs.
//   - (PathUtils.*): propagated from pathutils.PathOsToCfs.
func ListFiles(cfsPath *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	dir := &pathutils.PathCfs{Value: "code-from-spec/functional/logic"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		if errors.Is(err, listfiles.ErrDirectoryNotFound) {
			log.Fatal("directory does not exist")
		}
		if errors.Is(err, listfiles.ErrWalkError) {
			log.Fatal("error walking directory tree")
		}
		// PathUtils errors propagated from PathCfsToOs or PathOsToCfs.
		if errors.Is(err, pathutils.ErrPathEmpty) {
			log.Fatal("path must not be empty")
		}
		if errors.Is(err, pathutils.ErrDirectoryTraversal) {
			log.Fatal("path traversal detected")
		}
		if errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			log.Fatal("path resolves outside project root")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	fmt.Printf("found %d file(s):\n", len(files))
	for _, f := range files {
		fmt.Println(" ", f.Value)
	}
}
```
