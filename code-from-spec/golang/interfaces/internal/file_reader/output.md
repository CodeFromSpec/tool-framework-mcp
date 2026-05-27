[//]: # (code-from-spec: ROOT/golang/internal/file_reader/interface@Evuc5p1KT2M09b5PC1NLEfHsxXo)

# `filereader` Package — Go Interface Specification

Package `filereader` provides sequential line-by-line reading of text files.

## Struct Definitions

```go
package filereader

// FileReader holds the state for reading a file line by line.
type FileReader struct {
	filePath string
}
```

## Error Sentinels

```go
import "errors"

// ErrEndOfFile is returned by ReadLine and SkipLines when no more lines
// are available, or after Close has been called.
var ErrEndOfFile = errors.New("end of file")

// ErrFileUnreadable is returned by OpenFileReader when the file cannot
// be opened.
var ErrFileUnreadable = errors.New("file unreadable")
```

## Function and Method Signatures

```go
// OpenFileReader opens the file at filePath and prepares it for
// sequential line-by-line reading. The file remains open until Close
// is called. Returns ErrFileUnreadable if the file cannot be opened.
func OpenFileReader(filePath string) (*FileReader, error)

// ReadLine reads the next line from the file. CRLF sequences are
// normalized to LF, and the line terminator is stripped before the
// line is returned. Returns ErrEndOfFile when no more lines are
// available.
func (r *FileReader) ReadLine() (string, error)

// SkipLines reads and discards count lines without returning their
// content. Returns ErrEndOfFile if the end of file is reached before
// count lines have been skipped.
func (r *FileReader) SkipLines(count int) error

// Close releases the file resource. After Close is called, any
// subsequent call to ReadLine or SkipLines returns ErrEndOfFile.
func (r *FileReader) Close()
```

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
)

func main() {
	reader, err := filereader.OpenFileReader("data.txt")
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer reader.Close()

	// Skip the first two header lines.
	if err := reader.SkipLines(2); err != nil && !errors.Is(err, filereader.ErrEndOfFile) {
		log.Fatalf("unexpected error skipping lines: %v", err)
	}

	// Read all remaining lines.
	for {
		line, err := reader.ReadLine()
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			log.Fatalf("unexpected error reading line: %v", err)
		}
		fmt.Println(line)
	}
}
```
