[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@v4p1ZQk6LI4SGm6vDNa_nUrgkrQ)

# Package `filereader`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader`

---

## Structs

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileReader holds the state for sequential line-by-line reading of a file.
// The caller must call FileClose when done to release the file handle.
type FileReader struct {
	CfsPath pathutils.PathCfs
}
```

---

## Error Sentinels

```go
package filereader

import "errors"

var ErrFileUnreadable = errors.New("file cannot be opened")
var ErrEndOfFile      = errors.New("end of file")
```

---

## Functions

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileOpen opens a file at cfsPath and prepares it for sequential
// line-by-line reading from the beginning of the file.
// The caller must call FileClose when done — failing to do so leaks the file handle.
// Returns ErrFileUnreadable if the file exists but cannot be opened.
// Propagates errors from pathutils.PathCfsToOs.
func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the terminator.
// Returns ErrEndOfFile when there are no more lines to read,
// or after FileClose has been called.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines without returning their content.
// Does nothing if FileClose has already been called.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource associated with reader.
// After FileClose, FileReadLine returns ErrEndOfFile and FileSkipLines does nothing.
func FileClose(reader *FileReader)
```

---

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
	cfsPath := &pathutils.PathCfs{Value: "EXTERNAL/data/sample.txt"}

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
