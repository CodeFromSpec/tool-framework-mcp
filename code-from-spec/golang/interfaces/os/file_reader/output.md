<!-- code-from-spec: ROOT/golang/interfaces/os/file_reader@zRSUqUXBgvJ1KgHMBz5yH2VbxEY -->

# Package `filereader`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
```

## Structs

```go
package filereader

// FileReader holds the state for sequential line-by-line reading of a file.
// It is opened via FileOpen and must be closed with FileClose when no longer
// needed to avoid leaking the file handle.
type FileReader struct {
	CfsPath *pathutils.PathCfs
}
```

## Error Sentinels

```go
package filereader

import "errors"

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
package filereader

// FileOpen opens the file at cfs_path and prepares it for sequential
// line-by-line reading from the beginning of the file. The caller must
// call FileClose when done — failing to do so leaks the file handle.
//
// Possible errors: ErrFileUnreadable, and any path error propagated from
// pathutils.PathCfsToOs (ErrPathEmpty, ErrPathAbsolute,
// ErrPathContainsBackslash, ErrDirectoryTraversal, ErrResolvesOutsideRoot,
// ErrCannotDetermineRoot).
func FileOpen(cfs_path *pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line content without the line terminator.
//
// Returns ErrEndOfFile when there are no more lines to read, or after
// FileClose has been called.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines from the file without
// returning their content. Does nothing if the reader has been closed.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource associated with reader. After
// FileClose is called, FileReadLine returns ErrEndOfFile and FileSkipLines
// does nothing.
func FileClose(reader *FileReader)
```

## Usage Example

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
	cfs := &pathutils.PathCfs{Value: "internal/filereader/filereader.go"}

	reader, err := filereader.FileOpen(cfs)
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer filereader.FileClose(reader)

	// Skip the first line.
	filereader.FileSkipLines(reader, 1)

	// Read lines until end of file.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			log.Fatalf("read error: %v", err)
		}
		fmt.Println(line)
	}
}
```
