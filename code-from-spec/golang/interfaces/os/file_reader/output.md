[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_reader@VeJdAr1QhIMFKgaT17LjPlWVbqI)

# Interface: `filereader`

## Package

```go
package filereader
```

## Structs

```go
// FileReader holds the state for sequential line-by-line reading of a file.
// Open a FileReader with FileOpen and release it with FileClose when done.
type FileReader struct {
	CfsPath pathutils.PathCfs
}
```

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

## Functions

```go
// FileOpen opens the file at cfs_path and prepares it for sequential
// line-by-line reading, starting from the beginning of the file.
// The caller must call FileClose when done — failing to do so leaks
// the file handle.
//
// Possible errors:
//   - ErrPathEmpty, ErrPathAbsolute, ErrPathContainsBackslash,
//     ErrDirectoryTraversal, ErrResolvesOutsideRoot: propagated from
//     pathutils.PathCfsToOs if the path is invalid.
//   - ErrFileUnreadable: the path is valid but the file cannot be opened.
func FileOpen(cfs_path pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the line terminator.
//
// Possible errors:
//   - ErrEndOfFile: no more lines to read, or the reader has been closed.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines without returning their
// content. If fewer than count lines remain, it reads until end of file
// without returning an error.
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

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/os/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/os/pathutils"
)

func main() {
	// Open a file for sequential reading.
	cfs := pathutils.PathCfs{Value: "code-from-spec/golang/interfaces/os/file_reader/output.md"}
	r, err := filereader.FileOpen(cfs)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer filereader.FileClose(r)

	// Skip the first two lines.
	filereader.FileSkipLines(r, 2)

	// Read remaining lines one at a time.
	for {
		line, err := filereader.FileReadLine(r)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
		fmt.Println(line)
	}
}
```
