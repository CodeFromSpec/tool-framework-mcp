// code-from-spec: ROOT/golang/internal/file_reader/tests@BDnomavKsL_lKqZDClBFYcX6QAc

// Package filereader provides tests for the FileReader type.
// These tests are written as internal tests (same package) so they can verify
// behaviour through the public interface without any test framework beyond the
// standard "testing" package.
package filereader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// testWriteFile creates a file in dir with the given name and raw byte content.
// It calls t.Fatal on any I/O error, so callers do not need to check the
// return value.
func testWriteFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("testWriteFile: could not write %q: %v", path, err)
	}
	return path
}

// ----------------------------------------------------------------------------
// Happy Path Tests
// ----------------------------------------------------------------------------

// TestOpensAndReadsAllLines verifies that a file with multiple LF-terminated
// lines is read correctly, one line at a time, and that ErrEndOfFile is
// returned after the last line.
func TestOpensAndReadsAllLines(t *testing.T) {
	dir := t.TempDir()
	// Three lines, LF endings.
	path := testWriteFile(t, dir, "lines.txt", []byte("first\nsecond\nthird\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}
	defer r.Close()

	// Table of expected lines in order.
	expected := []string{"first", "second", "third"}
	for i, want := range expected {
		got, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine[%d]: unexpected error: %v", i, err)
		}
		if got != want {
			t.Errorf("ReadLine[%d]: got %q, want %q", i, got, want)
		}
	}

	// After all lines, the next ReadLine must return ErrEndOfFile.
	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after last line: got %v, want ErrEndOfFile", err)
	}
}

// TestNormalizesCRLF verifies that CRLF line endings are normalised so that
// the returned line strings contain neither CR nor LF characters.
func TestNormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	// Three lines with CRLF endings.
	path := testWriteFile(t, dir, "crlf.txt", []byte("alpha\r\nbeta\r\ngamma\r\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}
	defer r.Close()

	expected := []string{"alpha", "beta", "gamma"}
	for i, want := range expected {
		got, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine[%d]: unexpected error: %v", i, err)
		}
		if got != want {
			t.Errorf("ReadLine[%d]: got %q, want %q", i, got, want)
		}
		// Explicitly verify no stray CR or LF in the returned value.
		for _, ch := range got {
			if ch == '\r' || ch == '\n' {
				t.Errorf("ReadLine[%d]: returned line contains CR or LF: %q", i, got)
			}
		}
	}
}

// TestReadsFileWithNoTrailingNewline verifies that when the last line of a
// file has no trailing newline, it is still returned correctly and the
// subsequent ReadLine returns ErrEndOfFile.
func TestReadsFileWithNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	// Last line "c" has no newline.
	path := testWriteFile(t, dir, "no_trail.txt", []byte("a\nb\nc"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}
	defer r.Close()

	expected := []string{"a", "b", "c"}
	for i, want := range expected {
		got, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine[%d]: unexpected error: %v", i, err)
		}
		if got != want {
			t.Errorf("ReadLine[%d]: got %q, want %q", i, got, want)
		}
	}

	// After reading "c" (the unterminated last line), the next call must
	// signal ErrEndOfFile.
	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after last (unterminated) line: got %v, want ErrEndOfFile", err)
	}
}

// TestSkipLinesAdvancesReader verifies that SkipLines(2) skips exactly 2
// lines so that the next ReadLine returns the third line.
func TestSkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	// 5 lines.
	path := testWriteFile(t, dir, "five.txt", []byte("line1\nline2\nline3\nline4\nline5\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}
	defer r.Close()

	// Skip the first 2 lines; the next ReadLine should return "line3".
	r.SkipLines(2)

	got, err := r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine after SkipLines(2): unexpected error: %v", err)
	}
	const want = "line3"
	if got != want {
		t.Errorf("ReadLine after SkipLines(2): got %q, want %q", got, want)
	}
}

// TestSkipLinesPastEndOfFile verifies that calling SkipLines with a count
// greater than the number of remaining lines does not return an error and
// that the subsequent ReadLine returns ErrEndOfFile.
func TestSkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	// 2 lines.
	path := testWriteFile(t, dir, "two.txt", []byte("line1\nline2\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}
	defer r.Close()

	// Skipping 10 lines on a 2-line file must not panic or return an error.
	// SkipLines has no return value, so we just call it and verify that
	// ReadLine then returns ErrEndOfFile.
	r.SkipLines(10)

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after SkipLines past EOF: got %v, want ErrEndOfFile", err)
	}
}

// ----------------------------------------------------------------------------
// Edge Case Tests
// ----------------------------------------------------------------------------

// TestEmptyFile verifies that opening an empty file and immediately calling
// ReadLine returns ErrEndOfFile.
func TestEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "empty.txt", []byte{})

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}
	defer r.Close()

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine on empty file: got %v, want ErrEndOfFile", err)
	}
}

// TestSingleLineWithoutNewline verifies that a file containing a single word
// without any trailing newline is read correctly.
func TestSingleLineWithoutNewline(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "single.txt", []byte("hello"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}
	defer r.Close()

	got, err := r.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine: unexpected error: %v", err)
	}
	const want = "hello"
	if got != want {
		t.Errorf("ReadLine: got %q, want %q", got, want)
	}

	// The file is now exhausted; the next call must return ErrEndOfFile.
	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after single line: got %v, want ErrEndOfFile", err)
	}
}

// ----------------------------------------------------------------------------
// Failure Case Tests
// ----------------------------------------------------------------------------

// TestFileDoesNotExist verifies that calling OpenFileReader with a path that
// does not exist returns an error that satisfies errors.Is(err, ErrOpen).
func TestFileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	// Deliberately use a path that was never created.
	nonExistent := filepath.Join(dir, "does_not_exist.txt")

	_, err := OpenFileReader(nonExistent)
	if err == nil {
		t.Fatal("OpenFileReader: expected error, got nil")
	}
	if !errors.Is(err, ErrOpen) {
		t.Errorf("OpenFileReader: got %v, want errors.Is(err, ErrOpen) == true", err)
	}
}

// ----------------------------------------------------------------------------
// Close Behaviour Tests
// ----------------------------------------------------------------------------

// TestReadLineAfterClose verifies that calling ReadLine on a closed FileReader
// returns ErrEndOfFile, as required by the spec.
func TestReadLineAfterClose(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "close.txt", []byte("data\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}

	r.Close()

	_, err = r.ReadLine()
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("ReadLine after Close: got %v, want ErrEndOfFile", err)
	}
}

// TestDoubleCloseIsNoop verifies that calling Close a second time does not
// panic or cause any observable error (it must be a no-op).
func TestDoubleCloseIsNoop(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "double_close.txt", []byte("data\n"))

	r, err := OpenFileReader(path)
	if err != nil {
		t.Fatalf("OpenFileReader: unexpected error: %v", err)
	}

	r.Close()

	// A second Close must not panic. There is no return value to check.
	r.Close()
}
