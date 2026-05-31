// code-from-spec: ROOT/golang/implementation/os/file_reader@Gl2G0x7aeJj2yysFW1Map4_KK2U

package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrFileUnreadable is returned by FileOpen when the path is valid but
// the file cannot be opened (does not exist, permission denied, or
// other OS error).
var ErrFileUnreadable = errors.New("file unreadable")

// ErrEndOfFile is returned by FileReadLine when there are no more lines
// to read, including after FileClose has been called.
var ErrEndOfFile = errors.New("end of file")

// FileReader holds the state for reading a file line by line.
// Obtain one via FileOpen and release it with FileClose when done.
type FileReader struct {
	CfsPath *pathutils.PathCfs
	osPath  *pathutils.PathOs
	file    *os.File
	scanner *bufio.Scanner
	closed  bool
}

// FileOpen opens the file at cfsPath and prepares it for sequential
// line-by-line reading, starting from the beginning of the file.
// The caller must call FileClose when done — failing to do so leaks
// the file handle.
//
// Errors:
//   - ErrFileUnreadable: the path is valid but the file cannot be
//     opened (does not exist, permission denied, or other OS error).
//   - (PathUtils.*): propagated from PathCfsToOs.
func FileOpen(cfsPath *pathutils.PathCfs) (*FileReader, error) {
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(osPath.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFileUnreadable, err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(scanLinesRaw)

	reader := &FileReader{
		CfsPath: cfsPath,
		osPath:  osPath,
		file:    f,
		scanner: scanner,
		closed:  false,
	}

	return reader, nil
}

// FileReadLine reads the next line from the file, normalizes CRLF to
// LF, and returns the line without the terminator.
//
// Errors:
//   - ErrEndOfFile: there are no more lines to read, or the reader
//     has been closed.
func FileReadLine(reader *FileReader) (string, error) {
	if reader.closed {
		return "", ErrEndOfFile
	}

	if !reader.scanner.Scan() {
		return "", ErrEndOfFile
	}

	line := reader.scanner.Text()
	return line, nil
}

// FileSkipLines reads and discards count lines without returning their
// content. If the file has fewer than count lines remaining, it reads
// to the end and stops silently. After FileClose, FileSkipLines does
// nothing.
func FileSkipLines(reader *FileReader, count int) {
	if reader.closed {
		return
	}

	for i := 0; i < count; i++ {
		_, err := FileReadLine(reader)
		if err != nil {
			if errors.Is(err, ErrEndOfFile) {
				return
			}
			return
		}
	}
}

// FileClose releases the file resource held by reader. After FileClose,
// FileReadLine returns ErrEndOfFile and FileSkipLines does nothing.
func FileClose(reader *FileReader) {
	if reader.closed {
		return
	}

	reader.file.Close()
	reader.closed = true
}

// scanLinesRaw is a bufio.SplitFunc that reads lines and strips their
// terminators, handling both LF (\n) and CRLF (\r\n).
func scanLinesRaw(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	for i, b := range data {
		if b == '\n' {
			line := data[:i]
			// Strip trailing \r for CRLF sequences.
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			// Return a copy so the underlying buffer can be reused.
			result := make([]byte, len(line))
			copy(result, line)
			return i + 1, result, nil
		}
	}

	// At end of file with no newline — return remaining data as the last line.
	if atEOF {
		line := data
		// Strip trailing \r if present.
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		result := make([]byte, len(line))
		copy(result, line)
		return len(data), result, nil
	}

	// Request more data.
	return 0, nil, nil
}
