// code-from-spec: ROOT/golang/internal/file_reader/tests@TzrufNiCUrJCbqye4B8r3S9mVcY
package filereader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// testWriteFile creates a file with the given content in dir and returns its path.
func testWriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
	return path
}

// TestOpenAndReadAllLines opens a file with multiple LF-terminated lines and
// reads all lines, verifying each and expecting ErrEndOfFile after the last.
func TestOpenAndReadAllLines(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "lines.txt", "first\nsecond\nthird\n")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	want := []string{"first", "second", "third"}
	for _, expected := range want {
		line, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: unexpected error: %v", err)
		}
		if line != expected {
			t.Errorf("ReadLine: got %q, want %q", line, expected)
		}
	}

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after last line: got %v, want ErrEndOfFile", err)
	}
}

// TestNormalizesCRLF creates a file with CRLF endings and verifies that
// ReadLine returns lines without CR or LF characters.
func TestNormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "crlf.txt", "alpha\r\nbeta\r\ngamma\r\n")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	want := []string{"alpha", "beta", "gamma"}
	for _, expected := range want {
		line, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: unexpected error: %v", err)
		}
		if line != expected {
			t.Errorf("ReadLine: got %q, want %q", line, expected)
		}
	}
}

// TestNoTrailingNewline verifies that a file whose last line has no trailing
// newline is read correctly, and the subsequent ReadLine returns ErrEndOfFile.
func TestNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "notrail.txt", "line1\nline2")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	want := []string{"line1", "line2"}
	for _, expected := range want {
		line, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: unexpected error: %v", err)
		}
		if line != expected {
			t.Errorf("ReadLine: got %q, want %q", line, expected)
		}
	}

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after last line: got %v, want ErrEndOfFile", err)
	}
}

// TestSkipLinesAdvancesReader creates a 5-line file, skips 2 lines, and
// verifies that the next ReadLine returns the third line.
func TestSkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "five.txt", "line1\nline2\nline3\nline4\nline5\n")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	if err := r.SkipLines(2); err != nil {
		t.Fatalf("SkipLines(2): %v", err)
	}

	line, err := r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine: unexpected error: %v", err)
	}
	if line != "line3" {
		t.Errorf("ReadLine after SkipLines(2): got %q, want %q", line, "line3")
	}
}

// TestSkipLinesPastEndOfFile creates a 2-line file, calls SkipLines(10),
// and expects no error, then ReadLine returns ErrEndOfFile.
func TestSkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "two.txt", "a\nb\n")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	// SkipLines past end should not return an error per spec.
	if err := r.SkipLines(10); err != nil {
		t.Fatalf("SkipLines(10): unexpected error: %v", err)
	}

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after SkipLines past EOF: got %v, want ErrEndOfFile", err)
	}
}

// TestEmptyFile creates an empty file and expects ReadLine to return
// ErrEndOfFile immediately.
func TestEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "empty.txt", "")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine on empty file: got %v, want ErrEndOfFile", err)
	}
}

// TestSingleLineWithoutNewline creates a file containing only "hello" with
// no newline, reads it, and expects ErrEndOfFile on the next ReadLine.
func TestSingleLineWithoutNewline(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "single.txt", "hello")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	line, err := r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine: unexpected error: %v", err)
	}
	if line != "hello" {
		t.Errorf("ReadLine: got %q, want %q", line, "hello")
	}

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after single line: got %v, want ErrEndOfFile", err)
	}
}

// TestFileDoesNotExist calls OpenFileReader with a non-existent path and
// expects ErrFileUnreadable.
func TestFileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	_, err := OpenFileReader(path)
	if !errors.Is(err, ErrFileUnreadable) {
		t.Errorf("OpenFileReader on missing file: got %v, want ErrFileUnreadable", err)
	}
}

// TestClosePreventsFurtherReads verifies that after Close is called,
// ReadLine and SkipLines both return ErrEndOfFile.
func TestClosePreventsFurtherReads(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "close.txt", "line1\nline2\n")

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}

	r.Close()

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after Close: got %v, want ErrEndOfFile", err)
	}

	err = r.SkipLines(1)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("SkipLines after Close: got %v, want ErrEndOfFile", err)
	}
}
