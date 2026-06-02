// code-from-spec: ROOT/golang/tests/os/file_reader@KL_hGODBPdBsLZKY-it4KaUjLG0
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

func TestFileOpen_ReadsAllLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\nbeta\ngamma\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "alpha" {
		t.Errorf("got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "beta" {
		t.Errorf("got %q, want %q", line, "beta")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "gamma" {
		t.Errorf("got %q, want %q", line, "gamma")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileOpen_NormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\r\nbeta\r\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "alpha" {
		t.Errorf("got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "beta" {
		t.Errorf("got %q, want %q", line, "beta")
	}
}

func TestFileOpen_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\nbeta"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "alpha" {
		t.Errorf("got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "beta" {
		t.Errorf("got %q, want %q", line, "beta")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLines_AdvancesReader(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
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

func TestFileSkipLines_PastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("one\ntwo\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
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

func TestFileReadLine_PreservesLeadingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("  alpha\n    beta\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "  alpha" {
		t.Errorf("got %q, want %q", line, "  alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "    beta" {
		t.Errorf("got %q, want %q", line, "    beta")
	}
}

func TestFileReadLine_PreservesTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha  \nbeta   \n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "alpha  " {
		t.Errorf("got %q, want %q", line, "alpha  ")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "beta   " {
		t.Errorf("got %q, want %q", line, "beta   ")
	}
}

func TestFileReadLine_PreservesInternalWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha   beta\none\ttwo\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "alpha   beta" {
		t.Errorf("got %q, want %q", line, "alpha   beta")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "one\ttwo" {
		t.Errorf("got %q, want %q", line, "one\ttwo")
	}
}

func TestFileReadLine_PreservesEmptyLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\n\n\nbeta\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "alpha" {
		t.Errorf("got %q, want %q", line, "alpha")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "" {
		t.Errorf("got %q, want empty string", line)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "" {
		t.Errorf("got %q, want empty string", line)
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "beta" {
		t.Errorf("got %q, want %q", line, "beta")
	}
}

func TestFileReadLine_PreservesNonASCII(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("café\n日本語\n🎉🚀\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "café" {
		t.Errorf("got %q, want %q", line, "café")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "日本語" {
		t.Errorf("got %q, want %q", line, "日本語")
	}

	line, err = filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "🎉🚀" {
		t.Errorf("got %q, want %q", line, "🎉🚀")
	}
}

func TestFileOpen_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileOpen_SingleLineNoNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	line, err := filereader.FileReadLine(reader)
	if err != nil {
		t.Fatalf("FileReadLine: %v", err)
	}
	if line != "hello" {
		t.Errorf("got %q, want %q", line, "hello")
	}

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileOpen_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := &pathutils.PathCfs{Value: "nonexistent/file.txt"}
	_, err := filereader.FileOpen(cfsPath)
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReadLine_AfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "file.txt"}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLines_AfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
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
