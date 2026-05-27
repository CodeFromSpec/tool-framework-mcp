// code-from-spec: ROOT/golang/internal/file_reader/code@t76IbPqNBuCGg9iqZSU0BeN7_HE
package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// FileReader holds the state for reading a file line by line.
type FileReader struct {
	filePath string
	file     *os.File
	scanner  *bufio.Scanner
	closed   bool
}

// ErrEndOfFile is returned by ReadLine and SkipLines when no more lines
// are available, or after Close has been called.
var ErrEndOfFile = errors.New("end of file")

// ErrFileUnreadable is returned by OpenFileReader when the file cannot
// be opened.
var ErrFileUnreadable = errors.New("file unreadable")

// OpenFileReader opens the file at filePath and prepares it for
// sequential line-by-line reading. The file remains open until Close
// is called. Returns ErrFileUnreadable if the file cannot be opened.
func OpenFileReader(filePath string) (*FileReader, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileUnreadable, filePath)
	}

	scanner := bufio.NewScanner(f)

	return &FileReader{
		filePath: filePath,
		file:     f,
		scanner:  scanner,
		closed:   false,
	}, nil
}

// ReadLine reads the next line from the file. CRLF sequences are
// normalized to LF, and the line terminator is stripped before the
// line is returned. Returns ErrEndOfFile when no more lines are
// available.
func (r *FileReader) ReadLine() (string, error) {
	if r.closed {
		return "", ErrEndOfFile
	}

	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", fmt.Errorf("%w: %s", ErrEndOfFile, err.Error())
		}
		return "", ErrEndOfFile
	}

	line := r.scanner.Text()
	line = strings.TrimRight(line, "\r")
	return line, nil
}

// SkipLines reads and discards count lines without returning their
// content. Returns ErrEndOfFile if the end of file is reached before
// count lines have been skipped.
func (r *FileReader) SkipLines(count int) error {
	if r.closed {
		return ErrEndOfFile
	}

	for i := 0; i < count; i++ {
		if !r.scanner.Scan() {
			if err := r.scanner.Err(); err != nil {
				return fmt.Errorf("%w: %s", ErrEndOfFile, err.Error())
			}
			return ErrEndOfFile
		}
	}

	return nil
}

// Close releases the file resource. After Close is called, any
// subsequent call to ReadLine or SkipLines returns ErrEndOfFile.
func (r *FileReader) Close() {
	if r.closed {
		return
	}

	_ = r.file.Close()
	r.closed = true
}
