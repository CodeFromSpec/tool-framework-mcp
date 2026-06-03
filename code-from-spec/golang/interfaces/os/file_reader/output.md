[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@os7sRZ84rFe2RDbTyOXnBJKEkBw)

# Package `filereader`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
```

## Structs

```go
package filereader

type FileReader struct {
	CfsPath pathutils.PathCfs
}
```

## Error Sentinels

```go
package filereader

import "errors"

var ErrFileUnreadable = errors.New("file unreadable")
var ErrEndOfFile = errors.New("end of file")
```

## Functions

```go
package filereader

// FileOpen opens the file at cfsPath and prepares it for sequential
// line-by-line reading. The caller must call FileClose when done.
func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the reader, normalizes CRLF to LF,
// and returns the line without the terminator.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines without returning their content.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource held by the reader.
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
	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/golang/interfaces/os/file_reader/output.md"}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer filereader.FileClose(reader)

	for {
		line, err := filereader.FileReadLine(reader)
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
