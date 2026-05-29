// code-from-spec: ROOT/golang/tests/os/file_reader@kS-o37iLa4E5gLpj3G_vi2W8zN4
package filereader_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

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

func testOpenFile(t *testing.T, name string, content []byte) *pathutils.PathCfs {
	t.Helper()
	if err := os.WriteFile(name, content, 0600); err != nil {
		t.Fatalf("testOpenFile: write %q: %v", name, err)
	}
	return &pathutils.PathCfs{Value: name}
}

// --- Happy Path ---

func TestOpensAndReadsAllLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\nbeta\ngamma\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	testCases := []string{"alpha", "beta", "gamma"}
	for _, want := range testCases {
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
		t.Errorf("FileReadLine after last line: got error %v, want ErrEndOfFile", err)
	}

	filereader.FileClose(reader)
}

func TestNormalizesCRLFToLF(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\r\nbeta\r\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	filereader.FileClose(reader)
}

func TestReadsFileWithNoTrailingNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\nbeta"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

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
		t.Errorf("FileReadLine after last line: got error %v, want ErrEndOfFile", err)
	}

	filereader.FileClose(reader)
}

func TestFileSkipLinesAdvancesReader(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileSkipLines(reader, 2)

	got, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: unexpected error: %v", err)
	}
	if got != "three" {
		t.Errorf("FileReadLine after skip = %q, want %q", got, "three")
	}

	filereader.FileClose(reader)
}

func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("one\ntwo\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileSkipLines(reader, 10)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after skip past EOF: got error %v, want ErrEndOfFile", err)
	}

	filereader.FileClose(reader)
}

func TestPreservesLeadingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("  alpha\n    beta\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	for _, want := range []string{"  alpha", "    beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	filereader.FileClose(reader)
}

func TestPreservesTrailingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha  \nbeta   \n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	for _, want := range []string{"alpha  ", "beta   "} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	filereader.FileClose(reader)
}

func TestPreservesInternalWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha   beta\none\ttwo\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	for _, want := range []string{"alpha   beta", "one\ttwo"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	filereader.FileClose(reader)
}

func TestPreservesEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\n\n\nbeta\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	for _, want := range []string{"alpha", "", "", "beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	filereader.FileClose(reader)
}

func TestPreservesNonASCIICharacters(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("café\n日本語\n🎉🚀\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	for _, want := range []string{"café", "日本語", "🎉🚀"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FileReadLine = %q, want %q", got, want)
		}
	}

	filereader.FileClose(reader)
}

// --- Edge Cases ---

func TestEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte{})

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine on empty file: got error %v, want ErrEndOfFile", err)
	}

	filereader.FileClose(reader)
}

func TestSingleLineWithoutNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("hello"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	got, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("FileReadLine = %q, want %q", got, "hello")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after last line: got error %v, want ErrEndOfFile", err)
	}

	filereader.FileClose(reader)
}

// --- Failure Cases ---

func TestFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: filepath.Join("nonexistent", "path", "file.txt")}

	_, err := filereader.FileOpen(cfsPath)
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("FileOpen on nonexistent file: got error %v, want ErrFileUnreadable", err)
	}
}

func TestReadAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine after close: got error %v, want ErrEndOfFile", err)
	}
}

func TestSkipAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	// FileSkipLines after close should do nothing — no panic, no error.
	filereader.FileSkipLines(reader, 1)
}
