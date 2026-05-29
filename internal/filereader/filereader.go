// code-from-spec: ROOT/golang/implementation/os/file_reader@r99XOcUFy48N-ZLWbMwKAN69hHg

package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// FileReader holds the state for sequential line-by-line reading of a file.
// Obtain a FileReader via FileOpen. The caller must call FileClose when done
// to release the underlying file handle.
type FileReader struct {
	CfsPath *pathutils.PathCfs
	osPath  *pathutils.PathOs
	file    *os.File
	scanner *bufio.Scanner
	closed  bool
}

var (
	// ErrEndOfFile is returned by FileReadLine when there are no more lines
	// to read, including after FileClose has been called.
	ErrEndOfFile = errors.New("end of file")

	// ErrFileUnreadable is returned by FileOpen when the path is valid but
	// the file cannot be opened.
	ErrFileUnreadable = errors.New("file unreadable")
)

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
func FileOpen(cfs_path *pathutils.PathCfs) (*FileReader, error) {
	osPath, err := pathutils.PathCfsToOs(cfs_path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	scanner := bufio.NewScanner(f)

	return &FileReader{
		CfsPath: cfs_path,
		osPath:  osPath,
		file:    f,
		scanner: scanner,
		closed:  false,
	}, nil
}

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line without the line terminator.
//
// Possible errors:
//   - ErrEndOfFile when there are no more lines to read, or after
//     FileClose has been called.
func FileReadLine(reader *FileReader) (string, error) {
	if reader.closed {
		return "", ErrEndOfFile
	}

	if !reader.scanner.Scan() {
		return "", ErrEndOfFile
	}

	line := reader.scanner.Text()
	line = strings.TrimRight(line, "\r")

	return line, nil
}

// FileSkipLines reads and discards count lines from the file without
// returning their content. Does nothing if FileClose has already been
// called.
func FileSkipLines(reader *FileReader, count int) {
	if reader.closed {
		return
	}

	for i := 0; i < count; i++ {
		_, err := FileReadLine(reader)
		if errors.Is(err, ErrEndOfFile) {
			return
		}
	}
}

// FileClose releases the file resource associated with reader. After
// FileClose is called, FileReadLine returns ErrEndOfFile and
// FileSkipLines does nothing.
func FileClose(reader *FileReader) {
	if reader.closed {
		return
	}

	reader.file.Close()
	reader.file = nil
	reader.scanner = nil
	reader.closed = true
}
