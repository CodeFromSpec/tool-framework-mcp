[//]: # (code-from-spec: ROOT/golang/interfaces/os/list_files@89rzHP1DHinmdDADEVdjSdmOPBk)

# Interface: `listfiles`

## Package

```go
package listfiles
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/listfiles"
```

---

## Error Sentinels

```go
var (
	// ErrDirectoryNotFound is returned when the given directory does not exist.
	ErrDirectoryNotFound = errors.New("directory not found")

	// ErrWalk is returned when a filesystem error occurs while traversing the directory.
	ErrWalk = errors.New("walk error")
)
```

---

## Functions

```go
// ListFiles returns all files (not directories) found recursively under the
// given directory, as a sorted list of PathCfs values.
//
// If the directory exists but contains no files, an empty list is returned.
//
// Possible errors:
//   - pathutils.ErrPathEmpty
//   - pathutils.ErrPathAbsolute
//   - pathutils.ErrPathContainsBackslash
//   - pathutils.ErrDirectoryTraversal
//   - pathutils.ErrResolvesOutsideRoot
//   - pathutils.ErrCannotDetermineRoot
//   - ErrDirectoryNotFound
//   - ErrWalk
func ListFiles(cfs_path *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

func main() {
	dir := &pathutils.PathCfs{Value: "code-from-spec"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Value)
	}
}
```
