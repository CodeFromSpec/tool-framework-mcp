[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_writer@Wx24KTsk-fUQ7dSxQxg6pDtTkB4)

# Package `filewriter`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
```

Package `filewriter` writes content to files at paths given in the framework's canonical path format (CFS). Intermediate directories are created as needed and existing files are overwritten.

---

## Error Sentinels

```go
package filewriter

import "errors"

// ErrCannotCreateDirectory is returned when an intermediate directory
// cannot be created while preparing to write the file.
var ErrCannotCreateDirectory = errors.New("cannot create directory")

// ErrCannotWriteFile is returned when the file cannot be written after
// the directory structure has been prepared.
var ErrCannotWriteFile = errors.New("cannot write file")
```

---

## Functions

```go
package filewriter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileWrite writes content to the file at cfsPath as UTF-8 encoded text.
// If the file already exists it is overwritten. If it does not exist it
// is created. Intermediate directories are created as needed.
//
// Content is written exactly as received — no normalization of line
// endings or other transformations is applied.
//
// The path is validated before any file or directory is created. If
// validation fails, no changes are made to the filesystem.
//
// Errors:
//   - ErrCannotCreateDirectory: an intermediate directory cannot be created.
//   - ErrCannotWriteFile: the file cannot be written.
//   - (PathUtils.*): propagated from PathCfsToOs.
func FileWrite(cfsPath *pathutils.PathCfs, content string) error
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	cfsPath := &pathutils.PathCfs{Value: "internal/output/result.txt"}

	content := "Hello, world!\n"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		if errors.Is(err, filewriter.ErrCannotCreateDirectory) {
			log.Fatal("failed to create intermediate directories")
		}
		if errors.Is(err, filewriter.ErrCannotWriteFile) {
			log.Fatal("failed to write file")
		}
		if errors.Is(err, pathutils.ErrPathEmpty) {
			log.Fatal("path must not be empty")
		}
		if errors.Is(err, pathutils.ErrDirectoryTraversal) {
			log.Fatal("path traversal detected")
		}
		if errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			log.Fatal("path escapes project root")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	fmt.Println("file written successfully")
}
```
