// code-from-spec: ROOT/golang/tests/os/file_reader@uOK5eFNQUr3Yie_9onCwQa8tMtc

package filereader_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// testWriteFile creates a file at the given relative path (from cwd) with the
// provided content, creating parent directories as needed.
func testWriteFile(t *testing.T, relPath string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(relPath), 0o755); err != nil {
		t.Fatalf("testWriteFile: mkdir: %v", err)
	}
	if err := os.WriteFile(relPath, content, 0o644); err != nil {
		t.Fatalf("testWriteFile: write: %v", err)
	}
}

// testChdir changes the working directory to dir and restores the original in
// cleanup.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: chdir: %v", err)
		}
	})
}

// testCfsPath returns a PathCfs with the given value.
func testCfsPath(value string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: value}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestFileReader_OpensAndReadsAllLines(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha\nbeta\ngamma\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "beta", "gamma"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after last line: got %v, want ErrEndOfFile", err)
	}
}

func TestFileReader_NormalizesCRLF(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha\r\nbeta\r\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}
}

func TestFileReader_ReadsFileWithNoTrailingNewline(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha\nbeta"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after last line: got %v, want ErrEndOfFile", err)
	}
}

func TestFileReader_FileSkipLinesAdvancesReader(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)

	got, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: unexpected error: %v", err)
	}
	if got != "three" {
		t.Errorf("FileReadLine = %q, want %q", got, "three")
	}
}

func TestFileReader_FileSkipLinesPastEndOfFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("one\ntwo\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	// Should not panic or return an error.
	filereader.FileSkipLines(reader, 10)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after skip past EOF: got %v, want ErrEndOfFile", err)
	}
}

func TestFileReader_PreservesLeadingWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("  alpha\n    beta\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"  alpha", "    beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesTrailingWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha  \nbeta   \n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha  ", "beta   "} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesInternalWhitespace(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha   beta\none\ttwo\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha   beta", "one\ttwo"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesEmptyLines(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha\n\n\nbeta\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "", "", "beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesNonASCIICharacters(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("café\n日本語\n🎉🚀\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"café", "日本語", "🎉🚀"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestFileReader_EmptyFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte{})

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine on empty file: got %v, want ErrEndOfFile", err)
	}
}

func TestFileReader_SingleLineWithoutNewline(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("hello"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	got, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("FileReadLine = %q, want %q", got, "hello")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after last line: got %v, want ErrEndOfFile", err)
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

func TestFileReader_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := filereader.FileOpen(testCfsPath("nonexistent.txt"))
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("FileOpen on missing file: got %v, want ErrFileUnreadable", err)
	}
}

func TestFileReader_ReadAfterClose(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after close: got %v, want ErrEndOfFile", err)
	}
}

func TestFileReader_SkipAfterClose(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "file.txt", []byte("alpha\n"))

	reader, err := filereader.FileOpen(testCfsPath("file.txt"))
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	// Should do nothing and not panic.
	filereader.FileSkipLines(reader, 1)
}
