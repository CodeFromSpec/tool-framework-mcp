// code-from-spec: ROOT/golang/tests/os/file_reader@ztrLoj1ruOqJmr-gvRlbAyBxE3g
package filereader_test

import (
	"errors"
	"os"
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

func testWriteFile(t *testing.T, name string, content []byte) {
	t.Helper()
	if err := os.WriteFile(name, content, 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func testOpen(t *testing.T, path string) *filereader.FileReader {
	t.Helper()
	cfsPath := &pathutils.PathCfs{Value: path}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	t.Cleanup(func() { filereader.FileClose(reader) })
	return reader
}

func TestOpensAndReadsAllLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta\ngamma\n"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha" {
		t.Errorf("expected alpha, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta" {
		t.Errorf("expected beta, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "gamma" {
		t.Errorf("expected gamma, got %q, err %v", line, err)
	}
	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestNormalizesCRLF(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\r\nbeta\r\n"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha" {
		t.Errorf("expected alpha, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta" {
		t.Errorf("expected beta, got %q, err %v", line, err)
	}
}

func TestReadsFileWithNoTrailingNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha" {
		t.Errorf("expected alpha, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta" {
		t.Errorf("expected beta, got %q, err %v", line, err)
	}
	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLinesAdvancesReader(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))

	reader := testOpen(t, "file.txt")

	filereader.FileSkipLines(reader, 2)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "three" {
		t.Errorf("expected three, got %q, err %v", line, err)
	}
}

func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\n"))

	reader := testOpen(t, "file.txt")

	filereader.FileSkipLines(reader, 10)

	_, err := filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestPreservesLeadingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("  alpha\n    beta\n"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "  alpha" {
		t.Errorf("expected '  alpha', got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "    beta" {
		t.Errorf("expected '    beta', got %q, err %v", line, err)
	}
}

func TestPreservesTrailingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha  \nbeta   \n"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha  " {
		t.Errorf("expected 'alpha  ', got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta   " {
		t.Errorf("expected 'beta   ', got %q, err %v", line, err)
	}
}

func TestPreservesInternalWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha   beta\none\ttwo\n"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha   beta" {
		t.Errorf("expected 'alpha   beta', got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "one\ttwo" {
		t.Errorf("expected 'one\\ttwo', got %q, err %v", line, err)
	}
}

func TestPreservesEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\n\n\nbeta\n"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha" {
		t.Errorf("expected alpha, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "" {
		t.Errorf("expected empty, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "" {
		t.Errorf("expected empty, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta" {
		t.Errorf("expected beta, got %q, err %v", line, err)
	}
}

func TestPreservesNonASCIICharacters(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("café\n日本語\n🎉🚀\n"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "café" {
		t.Errorf("expected café, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "日本語" {
		t.Errorf("expected 日本語, got %q, err %v", line, err)
	}
	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "🎉🚀" {
		t.Errorf("expected 🎉🚀, got %q, err %v", line, err)
	}
}

func TestEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte(""))

	reader := testOpen(t, "file.txt")

	_, err := filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestSingleLineWithoutNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("hello"))

	reader := testOpen(t, "file.txt")

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "hello" {
		t.Errorf("expected hello, got %q, err %v", line, err)
	}
	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "nonexistent/file.txt"}
	_, err := filereader.FileOpen(cfsPath)
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestReadAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile after close, got %v", err)
	}
}

func TestSkipAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 1)
}
