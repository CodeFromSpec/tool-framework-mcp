[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@yEtAaYmUrn7xyFFrwQUUdsv9SNI)

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

// FileReader holds the state for sequential line-by-line reading of a file.
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

// FileOpen opens the file at cfs_path and prepares it for sequential
// line-by-line reading from the beginning. The caller must call FileClose
// when done to avoid leaking the file handle.
func FileOpen(cfs_path *pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the reader, normalizes CRLF to LF,
// and returns the line without the line terminator. Returns ErrEndOfFile
// when no more lines are available.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines from the reader without
// returning their content.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource held by reader. After FileClose,
// FileReadLine returns ErrEndOfFile and FileSkipLines does nothing.
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
