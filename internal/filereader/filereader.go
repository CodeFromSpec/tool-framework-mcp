// code-from-spec: ROOT/golang/implementation/os/file_reader@MEeLfTrvL7FXmn5jy7rgq-xbH9o

package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// ErrEndOfFile is returned by FileReadLine when there are no more lines
// to read, including after FileClose has been called.
var ErrEndOfFile = errors.New("end of file")

// ErrFileUnreadable is returned by FileOpen when the path is valid but
// the file cannot be opened.
var ErrFileUnreadable = errors.New("file unreadable")

// FileReader holds the state for sequential line-by-line reading of a file.
// Obtain a FileReader via FileOpen. The caller must call FileClose when done
// to release the underlying file handle.
type FileReader struct {
	CfsPath *pathutils.PathCfs

	file    *os.File
	scanner *bufio.Scanner
	closed  bool
}

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
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, osPath.Value)
	}

	scanner := bufio.NewScanner(f)

	return &FileReader{
		CfsPath: cfs_path,
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

	// bufio.Scanner with the default ScanLines already strips line terminators,
	// but the spec requires explicit CRLF normalization. When using a custom
	// split function we handle it ourselves. With ScanLines, "\r" may still be
	// present at the end of lines on files with CRLF endings, so we strip it.
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
		if !reader.scanner.Scan() {
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
	reader.closed = true
	reader.file = nil
	reader.scanner = nil
}
