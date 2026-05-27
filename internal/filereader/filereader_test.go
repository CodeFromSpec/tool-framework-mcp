// code-from-spec: ROOT/golang/tests/internal/file_reader@gAwEDAzHhZAWZg1yqoNjhTRXPNM
package filereader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// testWriteFile creates a file at the given path with the given content.
func testWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// TestOpenAndReadAllLines verifies that OpenFileReader opens a file and
// ReadLine returns each line in order, followed by ErrEndOfFile.
func TestOpenAndReadAllLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	testWriteFile(t, path, []byte("alpha\nbeta\ngamma\n"))

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

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after last line: got %v, want ErrEndOfFile", err)
	}
}

// TestNormalizesCRLF verifies that CRLF line endings are normalized so
// the returned lines contain no CR or LF characters.
func TestNormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crlf.txt")
	testWriteFile(t, path, []byte("alpha\r\nbeta\r\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	cases := []string{"alpha", "beta"}
	for _, expected := range cases {
		line, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine: unexpected error: %v", err)
		}
		if line != expected {
			t.Errorf("ReadLine: got %q, want %q", line, expected)
		}
	}
}

// TestNoTrailingNewline verifies that a file whose last line has no
// trailing newline is read correctly.
func TestNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notrail.txt")
	testWriteFile(t, path, []byte("alpha\nbeta"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	line, err := r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine 1: unexpected error: %v", err)
	}
	if line != "alpha" {
		t.Errorf("ReadLine 1: got %q, want %q", line, "alpha")
	}

	line, err = r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine 2: unexpected error: %v", err)
	}
	if line != "beta" {
		t.Errorf("ReadLine 2: got %q, want %q", line, "beta")
	}

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine 3: got %v, want ErrEndOfFile", err)
	}
}

// TestSkipLinesAdvancesReader verifies that SkipLines skips the correct
// number of lines before the next ReadLine call.
func TestSkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "five.txt")
	testWriteFile(t, path, []byte("one\ntwo\nthree\nfour\nfive\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	if err := r.SkipLines(2); err != nil {
		t.Fatalf("SkipLines: unexpected error: %v", err)
	}

	line, err := r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine: unexpected error: %v", err)
	}
	if line != "three" {
		t.Errorf("ReadLine: got %q, want %q", line, "three")
	}
}

// TestSkipLinesPastEndOfFile verifies that SkipLines beyond the end of
// the file does not return an error, and the subsequent ReadLine returns
// ErrEndOfFile.
func TestSkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "two.txt")
	testWriteFile(t, path, []byte("one\ntwo\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	if err := r.SkipLines(10); err != nil {
		t.Fatalf("SkipLines: unexpected error: %v", err)
	}

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine: got %v, want ErrEndOfFile", err)
	}
}

// TestEmptyFile verifies that ReadLine on an empty file immediately
// returns ErrEndOfFile.
func TestEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	testWriteFile(t, path, []byte{})

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine: got %v, want ErrEndOfFile", err)
	}
}

// TestSingleLineWithoutNewline verifies that a file containing a single
// line with no newline is read correctly.
func TestSingleLineWithoutNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "single.txt")
	testWriteFile(t, path, []byte("hello"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}
	defer r.Close()

	line, err := r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine 1: unexpected error: %v", err)
	}
	if line != "hello" {
		t.Errorf("ReadLine 1: got %q, want %q", line, "hello")
	}

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine 2: got %v, want ErrEndOfFile", err)
	}
}

// TestFileDoesNotExist verifies that OpenFileReader returns ErrFileUnreadable
// when the target file does not exist.
func TestFileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	_, err := OpenFileReader(path)
	if !errors.Is(err, ErrFileUnreadable) {
		t.Errorf("OpenFileReader: got %v, want ErrFileUnreadable", err)
	}
}

// TestReadLineAfterClose verifies that ReadLine returns ErrEndOfFile after
// Close has been called.
func TestReadLineAfterClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "close.txt")
	testWriteFile(t, path, []byte("alpha\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}

	r.Close()

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after Close: got %v, want ErrEndOfFile", err)
	}
}

// TestSkipLinesAfterClose verifies that SkipLines returns ErrEndOfFile after
// Close has been called.
func TestSkipLinesAfterClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "close2.txt")
	testWriteFile(t, path, []byte("alpha\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: %v", err)
	}

	r.Close()

	if err := r.SkipLines(1); !errors.Is(err, ErrEndOfFile) {
		t.Errorf("SkipLines after Close: got %v, want ErrEndOfFile", err)
	}
}
