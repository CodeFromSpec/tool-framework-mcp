// code-from-spec: ROOT/golang/implementation/os/file_reader@K1YCQUQfCY5n_-uWK2CxR4t501s

package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// FileReader holds the state for sequential line-by-line reading of a file.
// It is opened via FileOpen and must be closed with FileClose when no longer
// needed to avoid leaking the file handle.
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
// line-by-line reading from the beginning of the file. The caller must
// call FileClose when done — failing to do so leaks the file handle.
//
// Possible errors: ErrFileUnreadable, and any path error propagated from
// pathutils.PathCfsToOs (ErrPathEmpty, ErrPathAbsolute,
// ErrPathContainsBackslash, ErrDirectoryTraversal, ErrResolvesOutsideRoot,
// ErrCannotDetermineRoot).
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

	return &FileReader{
		CfsPath: cfs_path,
		osPath:  osPath,
		file:    f,
		scanner: scanner,
		closed:  false,
	}, nil
}

// FileReadLine reads the next line from the file, normalizes CRLF to LF,
// and returns the line content without the line terminator.
//
// Returns ErrEndOfFile when there are no more lines to read, or after
// FileClose has been called.
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
// returning their content. Does nothing if the reader has been closed.
func FileSkipLines(reader *FileReader, count int) {
	if reader.closed {
		return
	}

	for i := 0; i < count; i++ {
		if !reader.scanner.Scan() {
			return
		}
	}
}

// FileClose releases the file resource associated with reader. After
// FileClose is called, FileReadLine returns ErrEndOfFile and FileSkipLines
// does nothing.
func FileClose(reader *FileReader) {
	if reader.closed {
		return
	}

	reader.file.Close()
	reader.closed = true
}
