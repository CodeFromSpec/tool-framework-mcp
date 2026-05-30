// code-from-spec: ROOT/golang/tests/os/file_reader@GcwA3c09D7q9tkgQ2PIL1XZCcQ0
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

// TestFileOpenAndReadAllLines verifies that FileOpen succeeds and all lines
// can be read sequentially until ErrEndOfFile is returned.
func TestFileOpenAndReadAllLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "beta")
	testReadExpect(t, reader, "gamma")
	testReadExpectEOF(t, reader)

	filereader.FileClose(reader)
}

// TestFileNormalizesCRLF verifies that CRLF line endings are normalized to LF
// and the returned strings contain no CR or LF characters.
func TestFileNormalizesCRLF(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\r\nbeta\r\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "beta")
}

// TestFileReadsNoTrailingNewline verifies that a file without a trailing newline
// is read correctly.
func TestFileReadsNoTrailingNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "beta")
	testReadExpectEOF(t, reader)
}

// TestFileSkipLinesAdvancesReader verifies that FileSkipLines advances the
// reader by the given count.
func TestFileSkipLinesAdvancesReader(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "one\ntwo\nthree\nfour\nfive\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)
	testReadExpect(t, reader, "three")
}

// TestFileSkipLinesPastEndOfFile verifies that FileSkipLines past end of file
// causes subsequent reads to return ErrEndOfFile.
func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "one\ntwo\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 10)
	testReadExpectEOF(t, reader)
}

// TestFilePreservesLeadingWhitespace verifies that leading spaces in lines
// are preserved.
func TestFilePreservesLeadingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "  alpha\n    beta\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "  alpha")
	testReadExpect(t, reader, "    beta")
}

// TestFilePreservesTrailingWhitespace verifies that trailing spaces in lines
// are preserved.
func TestFilePreservesTrailingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha  \nbeta   \n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha  ")
	testReadExpect(t, reader, "beta   ")
}

// TestFilePreservesInternalWhitespace verifies that internal spaces and tabs
// in lines are preserved.
func TestFilePreservesInternalWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha   beta\none\ttwo\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha   beta")
	testReadExpect(t, reader, "one\ttwo")
}

// TestFilePreservesEmptyLines verifies that empty lines are returned as empty
// strings and are not skipped.
func TestFilePreservesEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\n\n\nbeta\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "")
	testReadExpect(t, reader, "")
	testReadExpect(t, reader, "beta")
}

// TestFilePreservesNonASCII verifies that non-ASCII UTF-8 characters are
// preserved correctly.
func TestFilePreservesNonASCII(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "café\n日本語\n🎉🚀\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "café")
	testReadExpect(t, reader, "日本語")
	testReadExpect(t, reader, "🎉🚀")
}

// TestFileEmptyFile verifies that opening an empty file succeeds and
// FileReadLine immediately returns ErrEndOfFile.
func TestFileEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte{}, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpectEOF(t, reader)
}

// TestFileSingleLineWithoutNewline verifies that a file containing a single
// line with no trailing newline is read correctly.
func TestFileSingleLineWithoutNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte("hello"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "hello")
	testReadExpectEOF(t, reader)
}

// TestFileDoesNotExist verifies that FileOpen returns ErrFileUnreadable when
// the file does not exist.
func TestFileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "nonexistent/file.txt"}
	_, err := filereader.FileOpen(cfsPath)
	if err == nil {
		t.Fatal("FileOpen: expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Fatalf("FileOpen: expected ErrFileUnreadable, got: %v", err)
	}
}

// TestFileReadAfterClose verifies that FileReadLine returns ErrEndOfFile after
// FileClose has been called.
func TestFileReadAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}

	filereader.FileClose(reader)
	testReadExpectEOF(t, reader)
}

// TestFileSkipAfterClose verifies that FileSkipLines does nothing and does not
// panic after FileClose has been called.
func TestFileSkipAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: unexpected error: %v", err)
	}

	filereader.FileClose(reader)
	// FileSkipLines must do nothing and not panic after close.
	filereader.FileSkipLines(reader, 1)
}

// testReadExpect calls FileReadLine and asserts the returned line matches want.
func testReadExpect(t *testing.T, reader *filereader.FileReader, want string) {
	t.Helper()
	got, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("FileReadLine: got %q, want %q", got, want)
	}
}

// testReadExpectEOF calls FileReadLine and asserts it returns ErrEndOfFile.
func testReadExpectEOF(t *testing.T, reader *filereader.FileReader) {
	t.Helper()
	_, err := filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Fatalf("FileReadLine: expected ErrEndOfFile, got: %v", err)
	}
}
