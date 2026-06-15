// code-from-spec: SPEC/golang/tests/os/file_reader@4ArtTuLeDntVQOgQHELKLqn_eUY
package filereader_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
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

func TestFileOpenAndReadAllLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha\nbeta\ngamma\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

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

func TestFileReadLineNormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha\r\nbeta\r\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha" {
		t.Errorf("expected alpha without CR, got %q, err %v", line, err)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta" {
		t.Errorf("expected beta without CR, got %q, err %v", line, err)
	}
}

func TestFileReadLineNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha\nbeta"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

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
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "three" {
		t.Errorf("expected three, got %q, err %v", line, err)
	}
}

func TestFileSkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("one\ntwo\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
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
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("  alpha\n    beta\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "  alpha" {
		t.Errorf("expected '  alpha', got %q, err %v", line, err)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "    beta" {
		t.Errorf("expected '    beta', got %q, err %v", line, err)
	}
}

func TestFileReadLinePreservesTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha  \nbeta   \n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha  " {
		t.Errorf("expected 'alpha  ', got %q, err %v", line, err)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta   " {
		t.Errorf("expected 'beta   ', got %q, err %v", line, err)
	}
}

func TestFileReadLinePreservesInternalWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha   beta\none\ttwo\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha   beta" {
		t.Errorf("expected 'alpha   beta', got %q, err %v", line, err)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "one\ttwo" {
		t.Errorf("expected 'one\\ttwo', got %q, err %v", line, err)
	}
}

func TestFileReadLinePreservesEmptyLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha\n\n\nbeta\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "alpha" {
		t.Errorf("expected alpha, got %q, err %v", line, err)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "" {
		t.Errorf("expected empty string, got %q, err %v", line, err)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "" {
		t.Errorf("expected empty string, got %q, err %v", line, err)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil || line != "beta" {
		t.Errorf("expected beta, got %q, err %v", line, err)
	}
}

func TestFileReadLinePreservesNonASCII(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("café\n日本語\n🎉🚀\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

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

func TestFileOpenEmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte{}, 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
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
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil || line != "hello" {
		t.Errorf("expected hello, got %q, err %v", line, err)
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileOpenFileNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := filereader.FileOpen(pathutils.PathCfs{Value: "nonexistent/file.txt"})
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReadLineAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
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
	dir := t.TempDir()
	testChdir(t, dir)

	err := os.WriteFile("file.txt", []byte("alpha\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	reader, err := filereader.FileOpen(pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 1)
}
