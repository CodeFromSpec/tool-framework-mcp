[//]: # (code-from-spec: SPEC/golang/interfaces/os/file_writer@_w2irblBCxAZPkHGoypAH2jeK_I)

# Package `filewriter`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filewriter`

## Error Sentinels

```go
package filewriter

import "errors"

var ErrCannotCreateDirectory = errors.New("cannot create directory")
var ErrCannotWriteFile       = errors.New("cannot write file")
```

## Functions

```go
package filewriter

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FileWrite writes content to the file at cfsPath as UTF-8 encoded text.
// If the file exists, it is overwritten. If it does not exist, it is created.
// Intermediate directories are created as needed.
// Content is written exactly as received with no transformations.
// The path is validated before writing — if validation fails, no file or directory is created.
func FileWrite(cfsPath *pathutils.PathCfs, content string) error
```

## Usage Example

```go
package main

import (
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	cfsPath := &pathutils.PathCfs{Value: "output/generated/result.txt"}

	err := filewriter.FileWrite(cfsPath, "Hello, world!\n")
	if err != nil {
		log.Fatal(err)
	}
}
```
