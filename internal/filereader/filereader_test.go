// code-from-spec: ROOT/golang/tests/os/file_reader@ey7Rz_Ozp3gFb2gNklMlSFApQ1Y
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

func TestFileOpenAndReadLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "beta")
	testReadExpect(t, reader, "gamma")
	testReadExpectEOF(t, reader)
}

func TestFileReadLineNormalizesCRLF(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\r\nbeta\r\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "beta")
}

func TestFileReadLineNoTrailingNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\nbeta"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "beta")
	testReadExpectEOF(t, reader)
}

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
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)
	testReadExpect(t, reader, "three")
}

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
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 10)
	testReadExpectEOF(t, reader)
}

func TestFileReadLinePreservesLeadingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "  alpha\n    beta\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "  alpha")
	testReadExpect(t, reader, "    beta")
}

func TestFileReadLinePreservesTrailingWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha  \nbeta   \n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha  ")
	testReadExpect(t, reader, "beta   ")
}

func TestFileReadLinePreservesInternalWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha   beta\none\ttwo\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha   beta")
	testReadExpect(t, reader, "one\ttwo")
}

func TestFileReadLinePreservesEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\n\n\nbeta\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "alpha")
	testReadExpect(t, reader, "")
	testReadExpect(t, reader, "")
	testReadExpect(t, reader, "beta")
}

func TestFileReadLinePreservesNonASCII(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "café\n日本語\n🎉🚀\n"
	if err := os.WriteFile("file.txt", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "café")
	testReadExpect(t, reader, "日本語")
	testReadExpect(t, reader, "🎉🚀")
}

func TestFileOpenEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte(""), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpectEOF(t, reader)
}

func TestFileReadLineSingleLineNoNewline(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte("hello"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	testReadExpect(t, reader, "hello")
	testReadExpectEOF(t, reader)
}

func TestFileOpenNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "nonexistent/file.txt"}
	_, err := filereader.FileOpen(cfsPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Fatalf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReadLineAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)
	testReadExpectEOF(t, reader)
}

func TestFileSkipLinesAfterClose(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)
	filereader.FileSkipLines(reader, 1)
}

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

func testReadExpectEOF(t *testing.T, reader *filereader.FileReader) {
	t.Helper()
	_, err := filereader.FileReadLine(reader)
	if err == nil {
		t.Fatal("FileReadLine: expected ErrEndOfFile, got nil")
	}
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Fatalf("FileReadLine: expected ErrEndOfFile, got %v", err)
	}
}
