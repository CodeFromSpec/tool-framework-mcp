// code-from-spec: ROOT/golang/implementation/os/file_reader@rANRF-LoJBNYSjb5Tp-by0Prbi0
package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrEndOfFile is returned by FileReadLine when there are no more
// lines to read, or after FileClose has been called.
var ErrEndOfFile = errors.New("end of file")

// ErrFileUnreadable is returned by FileOpen when the path is valid
// but the file cannot be opened (does not exist, permission denied,
// or other OS error).
var ErrFileUnreadable = errors.New("file unreadable")

// FileReader holds the state for sequential line-by-line reading of a file.
// It is created by FileOpen and must be closed with FileClose when done.
type FileReader struct {
	// CfsPath is the CFS path of the file being read.
	CfsPath pathutils.PathCfs

	osPath  pathutils.PathOs
	file    *os.File
	scanner *bufio.Scanner
}

// FileOpen opens the file at cfs_path and prepares it for sequential
// line-by-line reading from the beginning of the file.
//
// The caller must call FileClose when done — failing to do so leaks
// the file handle.
//
// Returns an error if:
//   - path validation or conversion fails (errors propagated from
//     pathutils.PathCfsToOs, e.g. ErrPathIsEmpty, ErrDirectoryTraversal).
//   - the file cannot be opened (ErrFileUnreadable).
func FileOpen(cfs_path *pathutils.PathCfs) (*FileReader, error) {
	osPath, err := pathutils.PathCfsToOs(cfs_path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, osPath.Value)
	}

	scanner := bufio.NewScanner(f)

	reader := &FileReader{
		CfsPath: *cfs_path,
		osPath:  *osPath,
		file:    f,
		scanner: scanner,
	}

	return reader, nil
}

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the line terminator.
//
// Returns ErrEndOfFile when there are no more lines to read, or after
// FileClose has been called on the reader.
func FileReadLine(reader *FileReader) (string, error) {
	if reader.file == nil {
		return "", ErrEndOfFile
	}

	if !reader.scanner.Scan() {
		return "", ErrEndOfFile
	}

	line := reader.scanner.Text()
	return line, nil
}

// FileSkipLines reads and discards count lines from the file without
// returning their content.
//
// Does nothing if FileClose has already been called on the reader.
func FileSkipLines(reader *FileReader, count int) {
	if reader.file == nil {
		return
	}

	for i := 0; i < count; i++ {
		if !reader.scanner.Scan() {
			return
		}
	}
}

// FileClose releases the file resource associated with reader.
//
// After FileClose is called, FileReadLine returns ErrEndOfFile and
// FileSkipLines does nothing.
func FileClose(reader *FileReader) {
	if reader.file == nil {
		return
	}

	reader.file.Close()
	reader.file = nil
	reader.scanner = nil
}
