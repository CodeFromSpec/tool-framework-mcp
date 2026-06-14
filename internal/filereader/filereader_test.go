// code-from-spec: ROOT/golang/tests/os/file_reader@brbZBjtiyRT3QMKUaDQRrkt1JjU
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

func TestFileOpenReadsAllLines(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha\nbeta\ngamma\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	expected := []string{"alpha", "beta", "gamma"}
	for _, want := range expected {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLineNormalizesCRLF(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha\r\nbeta\r\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line1 != "alpha" {
		t.Errorf("got %q, want %q", line1, "alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line2 != "beta" {
		t.Errorf("got %q, want %q", line2, "beta")
	}
}

func TestFileReadLineNoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha\nbeta"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line1 != "alpha" {
		t.Errorf("got %q, want %q", line1, "alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line2 != "beta" {
		t.Errorf("got %q, want %q", line2, "beta")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLinesAdvancesReader(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)

	got, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "three" {
		t.Errorf("got %q, want %q", got, "three")
	}
}

func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("one\ntwo\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 10)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLinePreservesLeadingWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("  alpha\n    beta\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line1 != "  alpha" {
		t.Errorf("got %q, want %q", line1, "  alpha")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line2 != "    beta" {
		t.Errorf("got %q, want %q", line2, "    beta")
	}
}

func TestFileReadLinePreservesTrailingWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha  \nbeta   \n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line1 != "alpha  " {
		t.Errorf("got %q, want %q", line1, "alpha  ")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line2 != "beta   " {
		t.Errorf("got %q, want %q", line2, "beta   ")
	}
}

func TestFileReadLinePreservesInternalWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha   beta\none\ttwo\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line1, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line1 != "alpha   beta" {
		t.Errorf("got %q, want %q", line1, "alpha   beta")
	}

	line2, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line2 != "one\ttwo" {
		t.Errorf("got %q, want %q", line2, "one\ttwo")
	}
}

func TestFileReadLinePreservesEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha\n\n\nbeta\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	cases := []string{"alpha", "", "", "beta"}
	for i, want := range cases {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine[%d]: %v", i, err)
		}
		if got != want {
			t.Errorf("line %d: got %q, want %q", i, got, want)
		}
	}
}

func TestFileReadLinePreservesNonASCII(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("café\n日本語\n🎉🚀\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	expected := []string{"café", "日本語", "🎉🚀"}
	for i, want := range expected {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine[%d]: %v", i, err)
		}
		if got != want {
			t.Errorf("line %d: got %q, want %q", i, got, want)
		}
	}
}

func TestFileOpenEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReadLineSingleLineNoNewline(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	got, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileOpenFileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	_, err := filereader.FileOpen(&pathutils.PathCfs{Value: "nonexistent/file.txt"})
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReadLineAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile after close, got %v", err)
	}
}

func TestFileSkipLinesAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 1)
}
