[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_writer@iq5fdksJ2HRiIR02SSI5JqcbiP4)

# Interface: `filewriter`

## Package

```go
package filewriter
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filewriter"
```

---

## Error Sentinels

```go
var (
	// ErrCannotCreateDirectory is returned when an intermediate directory
	// cannot be created during a write operation.
	ErrCannotCreateDirectory = errors.New("cannot create directory")

	// ErrCannotWriteFile is returned when the file cannot be written.
	ErrCannotWriteFile = errors.New("cannot write file")
)
```

---

## Functions

```go
// FileWrite writes content to the file at cfs_path as UTF-8 encoded text.
// If the file exists, it is overwritten. If it does not exist, it is
// created. Intermediate directories are created as needed.
//
// Content is written exactly as received — no normalization of line
// endings or other transformations is applied.
//
// The path is validated before writing — if validation fails, no file
// or directory is created.
//
// Possible errors:
//   - pathutils.ErrPathEmpty
//   - pathutils.ErrPathAbsolute
//   - pathutils.ErrPathContainsBackslash
//   - pathutils.ErrDirectoryTraversal
//   - pathutils.ErrResolvesOutsideRoot
//   - pathutils.ErrCannotDetermineRoot
//   - ErrCannotCreateDirectory
//   - ErrCannotWriteFile
func FileWrite(cfs_path *pathutils.PathCfs, content string) error
```

---

## Usage Examples

```go
package main

import (
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

func main() {
	// Write content to a file, creating intermediate directories if needed.
	cfs := &pathutils.PathCfs{Value: "internal/output/result.txt"}
	err := filewriter.FileWrite(cfs, "Hello, world!\n")
	if err != nil {
		log.Fatal(err)
	}
}
```
