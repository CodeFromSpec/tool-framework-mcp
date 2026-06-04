[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_writer@LEyqN9xS7bLRWMlaQrI503t809A)

# Package `filewriter`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
```

## Error Sentinels

```go
package filewriter

import "errors"

var ErrCannotCreateDirectory = errors.New("cannot create directory")
var ErrCannotWriteFile = errors.New("cannot write file")
```

## Functions

```go
package filewriter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileWrite writes content to the file at cfs_path as UTF-8 encoded text.
// If the file exists, it is overwritten. If it does not exist, it is created.
// Intermediate directories are created as needed.
// The path is validated before writing — if validation fails, no file or directory is created.
func FileWrite(cfsPath *pathutils.PathCfs, content string) error
```

## Usage Example

```go
package main

import (
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	cfsPath := &pathutils.PathCfs{Value: "internal/output/result.txt"}

	err := filewriter.FileWrite(cfsPath, "hello, world\n")
	if err != nil {
		log.Fatal(err)
	}
}
```
