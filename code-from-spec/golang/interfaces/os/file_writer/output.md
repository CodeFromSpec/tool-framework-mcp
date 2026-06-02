[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_writer@iWl_1sQuco7BkZw2valSnyXFaQg)

# Package `filewriter`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"`

## Package Declaration

```go
package filewriter
```

## Error Sentinels

```go
package filewriter

import "errors"

var ErrCannotCreateDirectory = errors.New("cannot create directory")
var ErrCannotWriteFile       = errors.New("cannot write file")
```

## Function Signatures

```go
package filewriter

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

func FileWrite(cfs_path *pathutils.PathCfs, content string) error
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
	cfs := &pathutils.PathCfs{Value: "output/generated.go"}

	err := filewriter.FileWrite(cfs, "package main\n\nfunc main() {}\n")
	if err != nil {
		log.Fatal(err)
	}
}
```
