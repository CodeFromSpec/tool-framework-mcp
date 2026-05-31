// code-from-spec: ROOT/golang/tests/os/file_reader@NYcDvoy5JXAXYK8d6Q3pOc4scEQ
package filereader_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir for the duration of the test.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

// testWriteFile writes content to a relative path inside the working directory.
func testWriteFile(t *testing.T, name string, content []byte) {
	t.Helper()
	if err := os.WriteFile(name, content, 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// testPathCfs returns a PathCfs with the given forward-slash path.
func testPathCfs(value string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: value}
}

// TestFileReader_OpensAndReadsAllLines verifies sequential line reading.
func TestFileReader_OpensAndReadsAllLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta\ngamma\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha", "beta", "gamma"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReader_NormalizesCRLF verifies CRLF line endings are normalized to LF.
func TestFileReader_NormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\r\nbeta\r\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReader_NoTrailingNewline verifies reading a file without a trailing newline.
func TestFileReader_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReader_SkipLinesAdvancesReader verifies FileSkipLines advances past lines.
func TestFileReader_SkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	filereader.FileSkipLines(r, 2)

	got, err := filereader.FileReadLine(r)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "three" {
		t.Errorf("got %q, want %q", got, "three")
	}
}

// TestFileReader_SkipLinesPastEndOfFile verifies skipping past EOF causes no error.
func TestFileReader_SkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	filereader.FileSkipLines(r, 10)

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReader_PreservesLeadingWhitespace verifies leading spaces are not trimmed.
func TestFileReader_PreservesLeadingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("  alpha\n    beta\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	cases := []string{"  alpha", "    beta"}
	for _, want := range cases {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReader_PreservesTrailingWhitespace verifies trailing spaces are not trimmed.
func TestFileReader_PreservesTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha  \nbeta   \n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	cases := []string{"alpha  ", "beta   "}
	for _, want := range cases {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReader_PreservesInternalWhitespace verifies internal spaces and tabs are not altered.
func TestFileReader_PreservesInternalWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha   beta\none\ttwo\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	cases := []string{"alpha   beta", "one\ttwo"}
	for _, want := range cases {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReader_PreservesEmptyLines verifies empty lines are returned as empty strings.
func TestFileReader_PreservesEmptyLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\n\n\nbeta\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	cases := []string{"alpha", "", "", "beta"}
	for _, want := range cases {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReader_PreservesNonASCII verifies multibyte characters are returned unchanged.
func TestFileReader_PreservesNonASCII(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("café\n日本語\n🎉🚀\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	cases := []string{"café", "日本語", "🎉🚀"}
	for _, want := range cases {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// TestFileReader_EmptyFile verifies that opening an empty file yields ErrEndOfFile immediately.
func TestFileReader_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte{})

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReader_SingleLineWithoutNewline verifies a file with a single line and no newline.
func TestFileReader_SingleLineWithoutNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("hello"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(r)

	got, err := filereader.FileReadLine(r)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReader_FileDoesNotExist verifies ErrFileUnreadable is returned for a missing file.
func TestFileReader_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := filereader.FileOpen(testPathCfs("nonexistent/file.txt"))
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

// TestFileReader_ReadAfterClose verifies FileReadLine returns ErrEndOfFile after FileClose.
func TestFileReader_ReadAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(r)

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

// TestFileReader_SkipAfterClose verifies FileSkipLines does nothing after FileClose.
func TestFileReader_SkipAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\n"))

	r, err := filereader.FileOpen(testPathCfs("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(r)

	// Should not panic or error.
	filereader.FileSkipLines(r, 1)
}
