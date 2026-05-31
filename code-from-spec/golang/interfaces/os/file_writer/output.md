[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_writer@zAWkRhK27dB3ZidKaVKyCRThvCk)

# Package `filewriter`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
```

Provides file-writing utilities for the Code from Spec framework. All paths use the `PathCfs` format from the `pathutils` package.

---

## Error Sentinels

```go
package filewriter

import "errors"

// ErrCannotCreateDirectory is returned when an intermediate directory cannot be created.
var ErrCannotCreateDirectory = errors.New("cannot create directory")

// ErrCannotWriteFile is returned when the file cannot be written.
var ErrCannotWriteFile = errors.New("cannot write file")
```

---

## Functions

```go
package filewriter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileWrite writes content to the file at cfsPath as UTF-8 encoded text.
// If the file exists, it is overwritten. If it does not exist, it is created.
// Intermediate directories are created as needed.
//
// Content is written exactly as received — no normalization of line endings
// or other transformations.
//
// The path is validated before writing — if validation fails, no file or
// directory is created.
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
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	cfs := &pathutils.PathCfs{Value: "internal/output/result.txt"}

	err := filewriter.FileWrite(cfs, "Hello, world!\n")
	if err != nil {
		log.Fatalf("FileWrite: %v", err)
	}
}
```
