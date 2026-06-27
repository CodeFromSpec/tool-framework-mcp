[//]: # (code-from-spec: SPEC/golang/interfaces/os/file_reader@O_3_MOSQP3QTwfNGvvuud48XddA)

# Package `filereader`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader`

## Struct Definitions

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FileReader holds the state for sequential line-by-line reading of a file.
// The caller must call FileClose when done to release the underlying file handle.
type FileReader struct {
	CfsPath pathutils.PathCfs
}
```

## Error Sentinels

```go
package filereader

import "errors"

var ErrFileUnreadable = errors.New("file unreadable")
var ErrEndOfFile      = errors.New("end of file")
```

## Function Signatures

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FileOpen opens the file at cfsPath and prepares it for sequential
// line-by-line reading from the beginning. The caller must call FileClose
// when done — failing to do so leaks the file handle.
func FileOpen(cfsPath pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the reader, normalizes CRLF to LF,
// and returns the line without the line terminator. Returns ErrEndOfFile
// when there are no more lines, or after FileClose has been called.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines from the reader without
// returning their content. Does nothing if the reader has been closed.
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

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	cfsPath := pathutils.PathCfs{Value: "SPEC/myproject/some_spec.md"}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)

	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(line)
	}
}
```
