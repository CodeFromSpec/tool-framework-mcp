// code-from-spec: ROOT/golang/internal/file_reader/code@D-_iC-rnBl6K4XVw2_WTvrRgyBU

// Package filereader provides a forward-only, sequential, line-by-line file
// reader. Memory usage does not scale with file size because the file is read
// incrementally using a bufio.Scanner.
package filereader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Sentinel errors. All errors returned by this package wrap one of these so
// callers can match them with errors.Is().
var (
	// ErrOpen is returned when a file cannot be opened (does not exist,
	// permission denied, or any other I/O error).
	ErrOpen = errors.New("file unreadable")

	// ErrEndOfFile is returned by ReadLine when there are no more lines to
	// read, or when the reader has already been closed.
	ErrEndOfFile = errors.New("end of file")
)

// FileReader holds the state for sequential line reading of a single file.
// All fields are unexported; callers interact only through the public methods.
type FileReader struct {
	// filePath is the path of the file that was opened. Retained for
	// diagnostic purposes.
	filePath string

	// file is the underlying OS file handle. Set to nil after Close.
	file *os.File

	// scanner drives the incremental line reading. We use a bufio.Scanner
	// rather than bufio.Reader so that line splitting is handled for us.
	// The scanner is configured to split on lines (the default).
	scanner *bufio.Scanner

	// closed is true once Close has been called. ReadLine and SkipLines
	// check this flag before doing any real work.
	closed bool
}

// OpenFileReader opens the file at filePath and prepares it for sequential
// line-by-line reading. The caller must call Close when finished.
//
// Returns ErrOpen (wrapped) if the file cannot be opened for any reason.
func OpenFileReader(filePath string) (*FileReader, error) {
	f, err := os.Open(filePath)
	if err != nil {
		// Wrap the underlying OS error so the caller can still use
		// errors.Is(err, ErrOpen) while also having access to the detail.
		return nil, fmt.Errorf("%w: %s", ErrOpen, err)
	}

	// bufio.Scanner default split function is ScanLines, which handles
	// both LF and CRLF line endings. We still normalize manually in
	// ReadLine to be explicit and consistent with the spec.
	scanner := bufio.NewScanner(f)

	return &FileReader{
		filePath: filePath,
		file:     f,
		scanner:  scanner,
		closed:   false,
	}, nil
}

// ReadLine returns the next line from the file stream, without the line
// terminator. CRLF sequences are normalized before the terminator is stripped.
//
// Returns ErrEndOfFile (wrapped) when no more lines are available or when the
// reader has been closed.
func (r *FileReader) ReadLine() (string, error) {
	// If the reader has been closed, behave as if the stream is exhausted.
	if r.closed {
		return "", fmt.Errorf("%w", ErrEndOfFile)
	}

	// Advance the scanner to the next token (line). Scan() returns false
	// when the stream is exhausted or when a scanner error occurs.
	if !r.scanner.Scan() {
		// Check for a real scanner error first (e.g. a read error).
		// If scanner.Err() is nil the stream simply ended normally.
		if scanErr := r.scanner.Err(); scanErr != nil {
			// Treat read errors as end-of-file so callers see a consistent
			// sentinel, but include the underlying detail.
			return "", fmt.Errorf("%w: %s", ErrEndOfFile, scanErr)
		}
		return "", fmt.Errorf("%w", ErrEndOfFile)
	}

	// scanner.Text() already strips the trailing newline via ScanLines, but
	// the spec asks us to normalize explicitly. We apply the normalization
	// ourselves so that the behaviour is well-defined regardless of how the
	// scanner splits:
	//   1. Replace CRLF with LF.
	//   2. Strip a trailing LF.
	//   3. Strip a trailing bare CR.
	line := r.scanner.Text()
	line = strings.ReplaceAll(line, "\r\n", "\n")
	line = strings.TrimRight(line, "\n")
	line = strings.TrimRight(line, "\r")

	return line, nil
}

// SkipLines advances the reader by count lines without returning their content.
// Skipping past the end of the file is not an error; a subsequent ReadLine
// call will return ErrEndOfFile.
//
// If the reader has already been closed this is a no-op.
func (r *FileReader) SkipLines(count int) {
	// Honour the closed state: closed reader → do nothing.
	if r.closed {
		return
	}

	for i := 0; i < count; i++ {
		// Reuse the scanner directly to avoid allocating the returned
		// string. If Scan returns false the stream is exhausted; we stop
		// iterating without raising an error (matches the spec).
		if !r.scanner.Scan() {
			return
		}
	}
}

// Close releases the underlying file handle. Calling Close more than once is a
// no-op. After Close returns, ReadLine will return ErrEndOfFile and SkipLines
// will return immediately.
func (r *FileReader) Close() {
	// Guard against double-close: the spec explicitly requires it to be a
	// no-op and os.File.Close would return an error on a second call.
	if r.closed {
		return
	}

	r.closed = true

	// We ignore the error from file.Close() because the spec defines no
	// error return for Close, and there is nothing useful the caller can do
	// at this point.
	_ = r.file.Close()
}
