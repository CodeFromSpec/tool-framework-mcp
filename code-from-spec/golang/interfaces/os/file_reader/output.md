[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@P2BviCiAnq8_8Exvbvm--zkv8ik)

# Interface: `filereader`

## Package

```go
package filereader
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
```

---

## Struct Definitions

```go
// FileReader holds the state for sequential line-by-line reading of a file.
// Obtain a FileReader via FileOpen. The caller must call FileClose when done
// to release the underlying file handle.
type FileReader struct {
	CfsPath *pathutils.PathCfs
}
```

---

## Error Sentinels

```go
var (
	// ErrEndOfFile is returned by FileReadLine when there are no more lines
	// to read, including after FileClose has been called.
	ErrEndOfFile = errors.New("end of file")

	// ErrFileUnreadable is returned by FileOpen when the path is valid but
	// the file cannot be opened.
	ErrFileUnreadable = errors.New("file unreadable")
)
```

---

## Functions

```go
// FileOpen opens the file at cfs_path and prepares it for sequential
// line-by-line reading starting from the beginning of the file.
// The caller must call FileClose when done — failing to do so leaks
// the file handle.
//
// Possible errors:
//   - Path errors propagated from pathutils.PathCfsToOs if the path
//     is invalid (ErrPathEmpty, ErrPathAbsolute, ErrPathContainsBackslash,
//     ErrDirectoryTraversal, ErrResolvesOutsideRoot, ErrCannotDetermineRoot).
//   - ErrFileUnreadable if the path is valid but the file cannot be opened.
func FileOpen(cfs_path *pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the line terminator.
//
// Possible errors:
//   - ErrEndOfFile when there are no more lines to read, or after
//     FileClose has been called.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines from the file without
// returning their content. Does nothing if FileClose has already been
// called.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource associated with reader. After
// FileClose is called, FileReadLine returns ErrEndOfFile and
// FileSkipLines does nothing.
func FileClose(reader *FileReader)
```

---

## Usage Examples

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

func main() {
	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/os/file_reader/_node.md"}

	// Open the file for reading.
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer filereader.FileClose(reader)

	// Skip the first two lines.
	filereader.FileSkipLines(reader, 2)

	// Read all remaining lines until end of file.
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
