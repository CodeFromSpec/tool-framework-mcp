// code-from-spec: ROOT/golang/tests/os/file_reader@lHFgDBrDeYpToEp20ndZOqIrjDw
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

// testWriteFile writes content to path (relative to cwd), creating directories as needed.
func testWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// TC-01: Opens and reads all lines
func TestFileOpenAndReadAllLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta\ngamma\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "alpha" {
		t.Errorf("line 1: got %q, want %q", line1, "alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "beta" {
		t.Errorf("line 2: got %q, want %q", line2, "beta")
	}

	line3, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 3: %v", err)
	}
	if line3 != "gamma" {
		t.Errorf("line 3: got %q, want %q", line3, "gamma")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine 4: got err %v, want ErrEndOfFile", err)
	}
}

// TC-02: Normalizes CRLF to LF
func TestFileReadLineNormalizesCRLF(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\r\nbeta\r\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "alpha" {
		t.Errorf("line 1: got %q, want %q", line1, "alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "beta" {
		t.Errorf("line 2: got %q, want %q", line2, "beta")
	}
}

// TC-03: Reads file with no trailing newline
func TestFileReadLineNoTrailingNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "alpha" {
		t.Errorf("line 1: got %q, want %q", line1, "alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "beta" {
		t.Errorf("line 2: got %q, want %q", line2, "beta")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine 3: got err %v, want ErrEndOfFile", err)
	}
}

// TC-04: FileSkipLines advances the reader
func TestFileSkipLinesAdvancesReader(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "three" {
		t.Errorf("line 1: got %q, want %q", line1, "three")
	}
}

// TC-05: FileSkipLines past end of file
func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	// Should not panic or error
	filereader.FileSkipLines(reader, 10)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine: got err %v, want ErrEndOfFile", err)
	}
}

// TC-06: Preserves leading whitespace
func TestFileReadLinePreservesLeadingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("  alpha\n    beta\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "  alpha" {
		t.Errorf("line 1: got %q, want %q", line1, "  alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "    beta" {
		t.Errorf("line 2: got %q, want %q", line2, "    beta")
	}
}

// TC-07: Preserves trailing whitespace
func TestFileReadLinePreservesTrailingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha  \nbeta   \n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "alpha  " {
		t.Errorf("line 1: got %q, want %q", line1, "alpha  ")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "beta   " {
		t.Errorf("line 2: got %q, want %q", line2, "beta   ")
	}
}

// TC-08: Preserves internal whitespace
func TestFileReadLinePreservesInternalWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha   beta\none\ttwo\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "alpha   beta" {
		t.Errorf("line 1: got %q, want %q", line1, "alpha   beta")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "one\ttwo" {
		t.Errorf("line 2: got %q, want %q", line2, "one\ttwo")
	}
}

// TC-09: Preserves empty lines
func TestFileReadLinePreservesEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\n\n\nbeta\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "alpha" {
		t.Errorf("line 1: got %q, want %q", line1, "alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "" {
		t.Errorf("line 2: got %q, want %q", line2, "")
	}

	line3, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 3: %v", err)
	}
	if line3 != "" {
		t.Errorf("line 3: got %q, want %q", line3, "")
	}

	line4, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 4: %v", err)
	}
	if line4 != "beta" {
		t.Errorf("line 4: got %q, want %q", line4, "beta")
	}
}

// TC-10: Preserves non-ASCII characters
func TestFileReadLinePreservesNonASCII(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("café\n日本語\n🎉🚀\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "café" {
		t.Errorf("line 1: got %q, want %q", line1, "café")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 2: %v", err)
	}
	if line2 != "日本語" {
		t.Errorf("line 2: got %q, want %q", line2, "日本語")
	}

	line3, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 3: %v", err)
	}
	if line3 != "🎉🚀" {
		t.Errorf("line 3: got %q, want %q", line3, "🎉🚀")
	}
}

// TC-11: Empty file
func TestFileReadLineEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte{})

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine: got err %v, want ErrEndOfFile", err)
	}
}

// TC-12: Single line without newline
func TestFileReadLineSingleLineNoNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("hello"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine 1: %v", err)
	}
	if line1 != "hello" {
		t.Errorf("line 1: got %q, want %q", line1, "hello")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("FileReadLine 2: got err %v, want ErrEndOfFile", err)
	}
}

// TC-13: File does not exist
func TestFileOpenFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "nonexistent/file.txt"}
	_, err := filereader.FileOpen(cfsPath)
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("FileOpen: got err %v, want ErrFileUnreadable", err)
	}
}

// TC-14: Read after close
func TestFileReadLineAfterClose(t *testing.T) {
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
		t.Errorf("FileReadLine after close: got err %v, want ErrEndOfFile", err)
	}
}

// TC-15: Skip after close
func TestFileSkipLinesAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "file.txt", []byte("alpha\n"))

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	// Should not panic or error — does nothing
	filereader.FileSkipLines(reader, 1)
}
