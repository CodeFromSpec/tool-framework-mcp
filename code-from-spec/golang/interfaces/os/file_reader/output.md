[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@AC8uE0OBBVXRNUcxNVsS6_HzdHw)

# Package `filereader`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"`

## Package Declaration

```go
package filereader
```

## Struct Definitions

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

type FileReader struct {
	CfsPath pathutils.PathCfs
}
```

## Error Sentinels

```go
package filereader

import "errors"

var ErrFileUnreadable = errors.New("file cannot be opened")
var ErrEndOfFile      = errors.New("end of file")
```

## Function Signatures

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

func FileOpen(cfs_path *pathutils.PathCfs) (*FileReader, error)
func FileReadLine(reader *FileReader) (string, error)
func FileSkipLines(reader *FileReader, count int)
func FileClose(reader *FileReader)
```

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	cfs := &pathutils.PathCfs{Value: "code-from-spec/some-node/_node.md"}

	r, err := filereader.FileOpen(cfs)
	if err != nil {
		log.Fatal(err)
	}
	defer filereader.FileClose(r)

	for {
		line, err := filereader.FileReadLine(r)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			log.Fatal(err)
		}
		fmt.Println(line)
	}
}
```
