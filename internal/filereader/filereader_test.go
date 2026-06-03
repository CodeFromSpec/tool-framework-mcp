// code-from-spec: ROOT/golang/tests/os/file_reader@01LvdE_8T5Dmq7UYOW5MtE3UKVk
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
	if err := os.WriteFile(name, content, 0600); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestFileOpenAndReadAllLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta\ngamma\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "beta", "gamma"} {
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
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\r\nbeta\r\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLineNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\nbeta"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "beta"} {
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

func TestFileSkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"))

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
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("one\ntwo\n"))

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
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("  alpha\n    beta\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"  alpha", "    beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha  \nbeta   \n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha  ", "beta   "} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesInternalWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha   beta\none\ttwo\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha   beta", "one\ttwo"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesEmptyLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\n\n\nbeta\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"alpha", "", "", "beta"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReadLinePreservesNonASCII(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("café\n日本語\n🎉🚀\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer filereader.FileClose(reader)

	for _, want := range []string{"café", "日本語", "🎉🚀"} {
		got, err := filereader.FileReadLine(reader)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileOpenEmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte(""))

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
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("hello"))

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

func TestFileOpenNonExistentFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := filereader.FileOpen(&pathutils.PathCfs{Value: "nonexistent/file.txt"})
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReadLineAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	_, err = filereader.FileReadLine(reader)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileSkipLinesAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	testWriteFile(t, "file.txt", []byte("alpha\n"))

	reader, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 1)
}
