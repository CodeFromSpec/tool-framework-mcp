// code-from-spec: ROOT/golang/internal/file_reader/code@PENDING
package filereader

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Sentinel errors for FileReader operations.
var (
	ErrOpen      = errors.New("file unreadable")
	ErrEndOfFile = errors.New("end of file")
)

// FileReader provides sequential line-by-line access to a text file.
// Lines are produced with CRLF normalized to LF and terminators stripped.
type FileReader struct {
	filePath string
	lines    []string
	position int
}

// OpenFileReader reads the entire file into memory, normalizes CRLF to LF,
// and splits the content into lines. A trailing newline does not produce a
// phantom empty line.
func OpenFileReader(filePath string) (*FileReader, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrOpen, err)
	}

	text := string(data)

	// Normalize CRLF to LF.
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Split on LF.
	lines := strings.Split(text, "\n")

	// If the text ends with LF the split produces a trailing empty string;
	// remove it so a final newline does not create a phantom empty line.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return &FileReader{
		filePath: filePath,
		lines:    lines,
		position: 0,
	}, nil
}

// ReadLine returns the next line without a line terminator.
// Returns ErrEndOfFile when there are no more lines.
func (r *FileReader) ReadLine() (string, error) {
	if r.position >= len(r.lines) {
		return "", ErrEndOfFile
	}
	line := r.lines[r.position]
	r.position++
	return line, nil
}

// SkipLines advances the reader by count lines without returning their
// content. Skipping past end-of-file is not an error; the position is
// clamped to the total number of lines.
func (r *FileReader) SkipLines(count int) {
	r.position += count
	if r.position > len(r.lines) {
		r.position = len(r.lines)
	}
}
