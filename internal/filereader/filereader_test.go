// code-from-spec: ROOT/golang/tests/os/file_reader@gmZy8RKnA_gb9op_e6d6DSJKWVI

package filereader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// testMakePath creates a PathCfs pointing to the given filename relative to
// the current working directory.
func testMakePath(name string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: name}
}

// testWriteFile writes content to a file named name in the current working
// directory and returns its base name.
func testWriteFile(t *testing.T, name string, content string) string {
	t.Helper()
	err := os.WriteFile(name, []byte(content), 0600)
	if err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
	return name
}

// testChdir changes the working directory to dir and restores the original
// directory in t.Cleanup.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
}

// TestFileReaderOpensAndReadsAllLines verifies that FileOpen followed by
// repeated FileReadLine calls returns each line in order and then
// ErrEndOfFile.
func TestFileReaderOpensAndReadsAllLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha\nbeta\ngamma\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"alpha", "beta", "gamma"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReaderNormalizesCRLF verifies that CRLF line endings are
// normalized to LF (i.e. the returned string contains no CR).
func TestFileReaderNormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha\r\nbeta\r\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReaderNoTrailingNewline verifies that a file without a trailing
// newline still returns the last line correctly and then ErrEndOfFile.
func TestFileReaderNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha\nbeta")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileSkipLinesAdvancesReader verifies that FileSkipLines skips the
// correct number of lines and subsequent FileReadLine returns the right line.
func TestFileSkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "one\ntwo\nthree\nfour\nfive\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	FileSkipLines(reader, 2)

	got, err := FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: unexpected error: %v", err)
	}
	if got != "three" {
		t.Errorf("got %q, want %q", got, "three")
	}
}

// TestFileSkipLinesPastEndOfFile verifies that skipping more lines than the
// file contains does not error and subsequent FileReadLine returns ErrEndOfFile.
func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "one\ntwo\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	FileSkipLines(reader, 10) // more than the file contains — must not panic or error

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReaderPreservesLeadingWhitespace verifies that leading spaces in
// lines are returned unchanged.
func TestFileReaderPreservesLeadingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "  alpha\n    beta\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	cases := []string{"  alpha", "    beta"}
	for _, want := range cases {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReaderPreservesTrailingWhitespace verifies that trailing spaces in
// lines are returned unchanged.
func TestFileReaderPreservesTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha  \nbeta   \n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	cases := []string{"alpha  ", "beta   "}
	for _, want := range cases {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReaderPreservesInternalWhitespace verifies that internal spaces and
// tabs within a line are returned unchanged.
func TestFileReaderPreservesInternalWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha   beta\none\ttwo\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	cases := []string{"alpha   beta", "one\ttwo"}
	for _, want := range cases {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReaderPreservesEmptyLines verifies that empty lines in the file are
// returned as empty strings and are not skipped.
func TestFileReaderPreservesEmptyLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha\n\n\nbeta\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	cases := []string{"alpha", "", "", "beta"}
	for _, want := range cases {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReaderPreservesNonASCII verifies that UTF-8 multibyte characters
// pass through unchanged.
func TestFileReaderPreservesNonASCII(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "café\n日本語\n🎉🚀\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	cases := []string{"café", "日本語", "🎉🚀"}
	for _, want := range cases {
		got, err := FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReaderEmptyFile verifies that opening an empty file and immediately
// calling FileReadLine returns ErrEndOfFile.
func TestFileReaderEmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReaderSingleLineNoNewline verifies that a file containing a single
// line with no newline is read correctly.
func TestFileReaderSingleLineNoNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "hello")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer FileClose(reader)

	got, err := FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileOpenFileDoesNotExist verifies that FileOpen returns ErrFileUnreadable
// when the path does not exist on the filesystem.
func TestFileOpenFileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Do not create the file — use a path that is guaranteed not to exist.
	_, err := FileOpen(testMakePath("does_not_exist.txt"))
	if !errors.Is(err, ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

// TestFileReadLineAfterClose verifies that calling FileReadLine after
// FileClose returns ErrEndOfFile.
func TestFileReadLineAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	FileClose(reader)

	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileSkipLinesAfterClose verifies that calling FileSkipLines after
// FileClose does nothing and does not panic.
func TestFileSkipLinesAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.txt", "alpha\n")

	reader, err := FileOpen(testMakePath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	FileClose(reader)

	// Must not panic or error.
	FileSkipLines(reader, 1)

	// Subsequent read should still return ErrEndOfFile.
	_, err = FileReadLine(reader)
	if !errors.Is(err, ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// Ensure filepath is used (imported for potential future use in helpers).
var _ = filepath.Join
