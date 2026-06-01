[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@c6H34aL4b0kYLUgWK5hi45AxxZg)

# Package `filereader`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
```

Package `filereader` provides sequential line-by-line reading of files identified by framework-canonical (CFS) paths.

---

## Structs

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileReader holds the state for reading a file line by line.
// Obtain one via FileOpen and release it with FileClose when done.
type FileReader struct {
	CfsPath *pathutils.PathCfs
}
```

---

## Error Sentinels

```go
package filereader

import "errors"

// ErrFileUnreadable is returned by FileOpen when the path is valid but
// the file cannot be opened (does not exist, permission denied, or
// other OS error).
var ErrFileUnreadable = errors.New("file unreadable")

// ErrEndOfFile is returned by FileReadLine when there are no more lines
// to read, including after FileClose has been called.
var ErrEndOfFile = errors.New("end of file")
```

---

## Functions

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// FileOpen opens the file at cfsPath and prepares it for sequential
// line-by-line reading, starting from the beginning of the file.
// The caller must call FileClose when done — failing to do so leaks
// the file handle.
//
// Errors:
//   - ErrFileUnreadable: the path is valid but the file cannot be
//     opened (does not exist, permission denied, or other OS error).
//   - (PathUtils.*): propagated from PathCfsToOs.
func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the file, normalizes CRLF to
// LF, and returns the line without the terminator.
//
// Errors:
//   - ErrEndOfFile: there are no more lines to read, or the reader
//     has been closed.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines without returning their
// content. If the file has fewer than count lines remaining, it reads
// to the end and stops silently. After FileClose, FileSkipLines does
// nothing.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource held by reader. After FileClose,
// FileReadLine returns ErrEndOfFile and FileSkipLines does nothing.
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
	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/os/file_reader/_node.md"}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			log.Fatal("file could not be opened")
		}
		if errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			log.Fatal("path escapes project root")
		}
		log.Fatalf("open failed: %v", err)
	}
	defer filereader.FileClose(reader)

	// Skip the first two lines.
	filereader.FileSkipLines(reader, 2)

	// Read remaining lines until end of file.
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			log.Fatalf("read error: %v", err)
		}
		fmt.Println(line)
	}
}
```
