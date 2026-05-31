[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@LRZ1XwyGk1TtV72siaN9XCNQKn8)

# Package `filereader`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
```

Provides sequential line-by-line reading of files addressed by CFS-format paths.

---

## Structs

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileReader holds the state for reading a file line by line.
// Obtain a FileReader via FileOpen. The caller must call FileClose when done.
type FileReader struct {
	CfsPath *pathutils.PathCfs
}
```

---

## Error Sentinels

```go
package filereader

import "errors"

// ErrFileUnreadable is returned by FileOpen when the path is valid but the file
// cannot be opened (does not exist, permission denied, or other OS error).
var ErrFileUnreadable = errors.New("file unreadable")

// ErrEndOfFile is returned by FileReadLine when there are no more lines to read,
// or when called after FileClose.
var ErrEndOfFile = errors.New("end of file")
```

---

## Functions

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileOpen opens the file at cfsPath and prepares it for sequential
// line-by-line reading, starting from the beginning of the file.
// The caller must call FileClose when done — failing to do so leaks the file handle.
//
// Errors:
//   - ErrFileUnreadable: the path is valid but the file cannot be opened
//     (does not exist, permission denied, or other OS error).
//   - (PathUtils.*): propagated from PathCfsToOs.
func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the terminator.
//
// Errors:
//   - ErrEndOfFile: no more lines to read, or the reader has been closed.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines without returning their content.
// Does nothing if the reader has been closed.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource held by reader.
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
	cfs := &pathutils.PathCfs{Value: "internal/filereader/filereader.go"}

	r, err := filereader.FileOpen(cfs)
	if err != nil {
		log.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	// Skip the first two lines.
	filereader.FileSkipLines(r, 2)

	// Read remaining lines until end of file.
	for {
		line, err := filereader.FileReadLine(r)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			log.Fatalf("FileReadLine: %v", err)
		}
		fmt.Println(line)
	}
}
```
