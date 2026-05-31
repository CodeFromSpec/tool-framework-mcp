// code-from-spec: ROOT/golang/implementation/os/file_reader@RWw1TqOKj-eKtCyDZy7NCjxFaHM

package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrFileUnreadable is returned by FileOpen when the path is valid but the file
// cannot be opened (does not exist, permission denied, or other OS error).
var ErrFileUnreadable = errors.New("file unreadable")

// ErrEndOfFile is returned by FileReadLine when there are no more lines to read,
// or when called after FileClose.
var ErrEndOfFile = errors.New("end of file")

// FileReader holds the state for reading a file line by line.
// Obtain a FileReader via FileOpen. The caller must call FileClose when done.
type FileReader struct {
	CfsPath  *pathutils.PathCfs
	file     *os.File
	scanner  *bufio.Scanner
	isClosed bool
}

// FileOpen opens the file at cfsPath and prepares it for sequential
// line-by-line reading, starting from the beginning of the file.
// The caller must call FileClose when done — failing to do so leaks the file handle.
//
// Errors:
//   - ErrFileUnreadable: the path is valid but the file cannot be opened
//     (does not exist, permission denied, or other OS error).
//   - (PathUtils.*): propagated from PathCfsToOs.
func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error) {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	scanner := bufio.NewScanner(f)

	return &FileReader{
		CfsPath:  cfsPath,
		file:     f,
		scanner:  scanner,
		isClosed: false,
	}, nil
}

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the terminator.
//
// Errors:
//   - ErrEndOfFile: no more lines to read, or the reader has been closed.
func FileReadLine(reader *FileReader) (string, error) {
	if reader.isClosed {
		return "", ErrEndOfFile
	}

	if !reader.scanner.Scan() {
		return "", ErrEndOfFile
	}

	line := reader.scanner.Text()
	line = strings.TrimRight(line, "\r")

	return line, nil
}

// FileSkipLines reads and discards count lines without returning their content.
// Does nothing if the reader has been closed.
func FileSkipLines(reader *FileReader, count int) {
	if reader.isClosed {
		return
	}

	for i := 0; i < count; i++ {
		_, err := FileReadLine(reader)
		if errors.Is(err, ErrEndOfFile) {
			return
		}
	}
}

// FileClose releases the file resource held by reader.
// After FileClose, FileReadLine returns ErrEndOfFile and FileSkipLines does nothing.
func FileClose(reader *FileReader) {
	if reader.isClosed {
		return
	}

	reader.file.Close()
	reader.isClosed = true
}
