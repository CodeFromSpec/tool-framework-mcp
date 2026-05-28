// code-from-spec: ROOT/golang/tests/os/file_reader@1umvdWp7Rha53-0D0j0EuM9WC8s

package filereader_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
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
	if err := os.WriteFile(name, content, 0o644); err != nil {
		t.Fatalf("testOpenFile: write file: %v", err)
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
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "alpha" {
		t.Errorf("line 1: got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "beta" {
		t.Errorf("line 2: got %q, want %q", line, "beta")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 3: %v", err)
	}
	if line != "gamma" {
		t.Errorf("line 3: got %q, want %q", line, "gamma")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine 4: got %v, want ErrEndOfFile", err)
	}
}

func TestNormalizesCRLFToLF(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\r\nbeta\r\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "alpha" {
		t.Errorf("line 1: got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "beta" {
		t.Errorf("line 2: got %q, want %q", line, "beta")
	}
}

func TestReadsFileWithNoTrailingNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\nbeta"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "alpha" {
		t.Errorf("line 1: got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "beta" {
		t.Errorf("line 2: got %q, want %q", line, "beta")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine 3: got %v, want ErrEndOfFile", err)
	}
}

func TestFileSkipLinesAdvancesReader(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "three" {
		t.Errorf("got %q, want %q", line, "three")
	}
}

func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("one\ntwo\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 10)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine: got %v, want ErrEndOfFile", err)
	}
}

func TestPreservesLeadingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("  alpha\n    beta\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "  alpha" {
		t.Errorf("line 1: got %q, want %q", line, "  alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "    beta" {
		t.Errorf("line 2: got %q, want %q", line, "    beta")
	}
}

func TestPreservesTrailingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha  \nbeta   \n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "alpha  " {
		t.Errorf("line 1: got %q, want %q", line, "alpha  ")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "beta   " {
		t.Errorf("line 2: got %q, want %q", line, "beta   ")
	}
}

func TestPreservesInternalWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha   beta\none\ttwo\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "alpha   beta" {
		t.Errorf("line 1: got %q, want %q", line, "alpha   beta")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "one\ttwo" {
		t.Errorf("line 2: got %q, want %q", line, "one\ttwo")
	}
}

func TestPreservesEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("alpha\n\n\nbeta\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "alpha" {
		t.Errorf("line 1: got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "" {
		t.Errorf("line 2: got %q, want empty string", line)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 3: %v", err)
	}
	if line != "" {
		t.Errorf("line 3: got %q, want empty string", line)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 4: %v", err)
	}
	if line != "beta" {
		t.Errorf("line 4: got %q, want %q", line, "beta")
	}
}

func TestPreservesNonASCIICharacters(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("café\n日本語\n🎉🚀\n"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "café" {
		t.Errorf("line 1: got %q, want %q", line, "café")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line != "日本語" {
		t.Errorf("line 2: got %q, want %q", line, "日本語")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 3: %v", err)
	}
	if line != "🎉🚀" {
		t.Errorf("line 3: got %q, want %q", line, "🎉🚀")
	}
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
	defer filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine: got %v, want ErrEndOfFile", err)
	}
}

func TestSingleLineWithoutNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := testOpenFile(t, "file.txt", []byte("hello"))

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line != "hello" {
		t.Errorf("got %q, want %q", line, "hello")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine 2: got %v, want ErrEndOfFile", err)
	}
}

// --- Failure Cases ---

func TestFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "nonexistent_file.txt"}

	_, err := filereader.FileOpen(cfsPath)
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("FileOpen: got %v, want ErrFileUnreadable", err)
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
		t.Errorf("FileReadLine after close: got %v, want ErrEndOfFile", err)
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
